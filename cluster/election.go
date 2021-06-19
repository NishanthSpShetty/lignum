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

func sendConnectRequestLeader(client http.Client, host string, port int, requestBody []byte) error {
	leaderEndpoint := fmt.Sprintf("http://%s:%d%s", host, port, "/api/follower/register")
	resp, err := client.Post(leaderEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	response, err := ioutil.ReadAll(resp.Body)
	//TODO: We will just log whatever we receive from the leader for now
	log.Debug().Bytes("ConnectLeaderResponse", response).Send()
	return err
}

func connectToLeader(serviceKey string, clusteController ClusterController, requestBody []byte, client http.Client) {

	if !state.isLeader() {

		if !state.isConnectedLeader() {
			//loop if the current node becomes the leader
			log.Info().Msg("Registering this service as a follower to the cluster leader...")
			//get the leader information and send a follow request.
			leaderNode, err := clusteController.GetLeader(serviceKey)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}
			err = sendConnectRequestLeader(client, leaderNode.Host, leaderNode.Port, requestBody)

			if err != nil {
				log.Error().Err(err).Msg("Failed to register with the leader ")
			} else {
				//we are connected to leader,
				state.setConnectedToLeader(true)
				state.setLeaderNode(&leaderNode)
			}
		} else {
			//check if we can ping connected leader
			log.Debug().Msg("pinging leader")
			if !state.getLeader().Ping(client) {
				state.setConnectedToLeader(false)
			}

		}

	}
}

//FollowerRegistrationRoutine Connect this service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func FollowerRegistrationRoutine(ctx context.Context, appConfig config.Server, connectionInterval time.Duration, serviceId string, clusteController ClusterController) {

	thisNode := NewNode(serviceId, appConfig.Host, appConfig.Port)
	requestBody, _ := thisNode.Json()
	ticker := time.NewTicker(connectionInterval)

	client := http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: 5 * time.Millisecond,
	}
	go func() {
		log.Debug().Msg("starting leader follower routine")
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				connectToLeader(appConfig.ServiceKey, clusteController, requestBody, client)
			}
		}
	}()
}

func tryAcquireLock(node Node, c ClusterController, serviceKey string) (bool, error) {

	acquired, err := c.AcquireLock(node, serviceKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire lock")
		return false, err
	}

	if acquired {
		state.markLeader()
		log.Info().Msg("Lock acquired and marking the node as leader")
	}
	return acquired, nil
}

func leaderElection(ctx context.Context, node Node, c ClusterController, serviceKey string) {

	loggedOnce := false
	consulLockPingInterval := 10 * time.Millisecond
	ticker := time.NewTicker(consulLockPingInterval)
	//start polling to acquire the lock indefinitely
	acquired := state.isLeader()
	var err error
	for {
		select {
		case <-ticker.C:
			if acquired {
				ticker.Stop()
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
			ticker.Stop()
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
