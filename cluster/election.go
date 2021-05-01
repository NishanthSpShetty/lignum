package cluster

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

//isLeader mark it as true when the current node becomes the leader
var isLeader = false

//connectToLeader Connect this service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func ConnectToLeader(appConfig config.Server, serviceId string, clusteController ClusterController) {

	thisNode := NewNode(serviceId, appConfig.Host, appConfig.Port)
	requestBody, _ := thisNode.Json()
	for {
		//loop if the current node becomes the leader
		log.Info().Msg("Registering this service as a follower to the cluster leader...")
		//get the leader
		if !isLeader {
			leaderNode, err := clusteController.GetLeader(appConfig.ServiceKey)
			if err != nil {
				log.Error().Err(err).Send()
				//TODO: give it a second and loop back??
				time.Sleep(1 * time.Second)
				continue
			}
			//get the leader information and send a follow request.
			leaderEndpoint := fmt.Sprintf("http://%s:%d%s", leaderNode.Host, leaderNode.Port, "/service/api/follower/register")
			resp, err := http.Post(leaderEndpoint, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				log.Error().Err(err).Msg("Failed to register with the leader ")
				//FIXME : this is possible when the leader elections are still going on and we call GetLeader.
				// should i return or not?
				return
			}
			response, err := ioutil.ReadAll(resp.Body)
			log.Debug().Bytes("ConnectLeaderResponse", response).Send()
			break

			//send connect ping to leader
		} else {
			log.Info().Msg("Im the leader....")
			return
		}
	}
}

func leaderElection(node Node, c ClusterController, serviceKey string) {
	loggedOnce := false
	//start polling to aquire the lock indefinitely
	for {
		if isLeader {
			//if the current node is leader, stop the busy loop for now
			return
		}
		aquired, queryDuration, err := c.AquireLock(node, serviceKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to aquire lock")
			continue
		}
		if aquired {
			isLeader = aquired
			log.Info().Str("Took", queryDuration.String()).Msg("Lock aquired and marking the node  as leader")
		} else {

			if !loggedOnce {
				log.Debug().Msg("Lock is already taken, will check again...")
				loggedOnce = true
			}
			//every 10ms attempt to grab a lock on the consul.
			time.Sleep(10 * time.Millisecond)
		}

	}
}

func InitiateLeaderElection(serverConfig config.Server, nodeId string, c ClusterController) {

	go leaderElection(Node{
		Id:   nodeId,
		Host: serverConfig.Host,
		Port: serverConfig.Port,
	}, c, serverConfig.ServiceKey)
}
