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

func connectToLeader(serviceKey string, clusteController ClusterController, requestBody []byte, httpClient http.Client) {

	if !state.isLeader() {

		if !state.isConnectedLeader() {
			//loop if the current node becomes the leader
			log.Info().Msg("registering this service as a follower to the cluster leader.")
			//get the leader information and send a follow request.
			leaderNode, err := clusteController.GetLeader(serviceKey)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}
			err = sendConnectRequestLeader(httpClient, leaderNode.Host, leaderNode.Port, requestBody)

			if err != nil {
				log.Error().Err(err).Msg("failed to register with the leader ")
			} else {
				//we are connected to leader,
				state.setConnectedToLeader(true)
				state.setLeaderNode(&leaderNode)
			}
		} else {
			//check if we can ping connected leader
			if !state.getLeader().Ping(httpClient) {
				log.Error().Msg("unable to ping the leader, will query leader status again and register the follower")
				state.setConnectedToLeader(false)
			}
		}
	}
}

//FollowerRegistrationRoutine Connect this service as a follower to the elected leader.
//this will be running forever whenever there is a change in leader this routine will make sure to connect the follower to reelected service
func FollowerRegistrationRoutine(ctx context.Context, appConfig config.Config, serviceId string, clusteController ClusterController) {

	thisNode := NewNode(serviceId, appConfig.Server.Host, appConfig.Server.Port)
	requestBody := thisNode.Json()
	ticker := time.NewTicker(appConfig.Follower.RegistrationOrHealthCheckIntervalInSeconds)

	httpClient := http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: appConfig.Follower.HealthCheckTimeoutInMilliSeconds * time.Millisecond,
	}
	go func() {
		log.Debug().Msg("starting follower registration routine")
		for {
			select {
			case <-ctx.Done():

				log.Debug().Msg("stopping follower registration routine")
				ticker.Stop()
				return
			case <-ticker.C:
				connectToLeader(appConfig.Server.ServiceKey, clusteController, requestBody, httpClient)
			}
		}
	}()
}

func tryAcquireLock(node Node, c ClusterController, serviceKey string) (bool, error) {

	acquired, err := c.AcquireLock(node, serviceKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to acquire lock")
		return false, err
	}

	if acquired {
		state.markLeader()
		log.Info().Msg("lock acquired and marking the node as leader")
	}
	return acquired, nil
}

func leaderElection(ctx context.Context, node Node, c ClusterController, serviceKey string, electionInterval time.Duration) {

	loggedOnce := false
	ticker := time.NewTicker(electionInterval)
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
			log.Info().Err(ctx.Err()).Msg("shutting down leader election worker")
			ticker.Stop()
			return
		}
	}
}

func InitiateLeaderElection(ctx context.Context, appConfig config.Config, nodeId string, c ClusterController) {
	node := Node{
		Id:   nodeId,
		Host: appConfig.Server.Host,
		Port: appConfig.Server.Port,
	}

	tryAcquireLock(node, c, appConfig.Server.ServiceKey)

	go leaderElection(ctx, node, c, appConfig.Server.ServiceKey, appConfig.Consul.LeaderElectionIntervalInMilliSeconds)
}
