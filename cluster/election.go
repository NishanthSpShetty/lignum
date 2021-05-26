package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

//isLeader mark it as true when the current node becomes the leader
var isLeader = false

func sendConnectRequestLeader(host string, port int, requestBody []byte) error {
	leaderEndpoint := fmt.Sprintf("http://%s:%d%s", host, port, "/service/api/follower/register")
	resp, err := http.Post(leaderEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	response, err := ioutil.ReadAll(resp.Body)
	//TODO: We will just log whatever we recieve from the leader for now
	log.Debug().Bytes("ConnectLeaderResponse", response).Send()
	return err
}

//connectToLeader Connect this service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func ConnectToLeader(appConfig config.Server, serviceId string, clusteController ClusterController) {

	thisNode := NewNode(serviceId, appConfig.Host, appConfig.Port)
	requestBody, _ := thisNode.Json()
	for {
		if !isLeader {
			//loop if the current node becomes the leader
			log.Info().Msg("Registering this service as a follower to the cluster leader...")
			//get the leader information and send a follow request.
			leaderNode, err := clusteController.GetLeader(appConfig.ServiceKey)
			if err != nil {
				log.Error().Err(err).Send()
				//TODO: give it a second and loop back??
				time.Sleep(1 * time.Second)
				continue
			}
			err = sendConnectRequestLeader(leaderNode.Host, leaderNode.Port, requestBody)

			if err != nil {
				log.Error().Err(err).Msg("Failed to register with the leader ")
				//FIXME : this is possible when the leader elections are still going on and we call GetLeader.
				// should it return or not?
			} else {
				//registered successfully, return?
				return
			}
		} else {
			log.Info().Msg("Im the leader....")
			return
		}
	}
}

func tryAcquireLock(node Node, c ClusterController, serviceKey string) (bool, error) {

	acquired, err := c.AcquireLock(node, serviceKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire lock")
		return false, err
	}

	if acquired {
		isLeader = acquired
		log.Info().Msg("Lock acquired and marking the node as leader")
	}
	return acquired, nil
}

func leaderElection(ctx context.Context, node Node, c ClusterController, serviceKey string) {

	loggedOnce := false
	consulLockPingInterval := 10 * time.Millisecond
	ticker := time.NewTicker(consulLockPingInterval)
	//start polling to acquire the lock indefinitely
	acquired := isLeader
	var err error
	for {
		select {
		case <-ticker.C:
			if acquired {
				return
			}
			acquired, err = tryAcquireLock(node, c, serviceKey)

			if err != nil {
				log.Error().Err(err).Msg("failed to acquire lock, will check again")
			}

			if !loggedOnce {
				log.Debug().Msg("lock is already taken, will check again")
				loggedOnce = true
			}

		case <-ctx.Done():
			log.Info().Err(ctx.Err()).Msg("context closed, shutting down leader election worker")
			return
		}
	}
}

func InitiateLeaderElection(ctx context.Context, serverConfig config.Server, nodeId string, c ClusterController) {
	node := Node{
		Id:   nodeId,
		Host: serverConfig.Host,
		Port: serverConfig.Port,
	}

	tryAcquireLock(node, c, serverConfig.ServiceKey)

	go leaderElection(ctx, node, c, serverConfig.ServiceKey)
}
