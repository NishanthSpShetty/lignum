package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
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

func connectToLeader(serviceKey string, clusteController cluster_types.ClusterController, node types.Node, httpClient http.Client, msgStore *message.MessageStore) {

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

			//get the replication status of this node to send it to leader
			// 1. need all the topics in this node
			// 2. need current message offset per node.

			topics := msgStore.GetTopics()
			stat := make([]types.MessageStat, 0, len(topics))

			for _, t := range topics {
				stat = append(stat, types.MessageStat{
					Topic:  t.GetName(),
					Offset: t.GetCurrentOffset(),
				})

			}

			request := types.FollowerRegistration{
				Node:        node,
				MessageStat: stat,
			}

			requestBody, err := json.Marshal(request)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal request")
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
func FollowerRegistrationRoutine(ctx context.Context, appConfig config.Config, serviceId string, clusteController cluster_types.ClusterController, msgStore *message.MessageStore) {

	thisNode := cluster_types.NewNode(serviceId, appConfig.Server.Host, appConfig.Server.Port)
	ticker := time.NewTicker(appConfig.Follower.RegistrationOrLeaderCheckIntervalInSeconds)

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
				connectToLeader(appConfig.Server.ServiceKey, clusteController, thisNode, httpClient, msgStore)
			}
		}
	}()
}

func tryAcquireLock(node cluster_types.Node, c cluster_types.ClusterController, serviceKey string) (bool, error) {

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

func leaderElection(ctx context.Context, node cluster_types.Node, c cluster_types.ClusterController, serviceKey string, electionInterval time.Duration) {

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

func InitiateLeaderElection(ctx context.Context, appConfig config.Config, nodeId string, c cluster_types.ClusterController) {
	node := cluster_types.Node{
		Id:   nodeId,
		Host: appConfig.Server.Host,
		Port: appConfig.Server.Port,
	}

	tryAcquireLock(node, c, appConfig.Server.ServiceKey)

	go leaderElection(ctx, node, c, appConfig.Server.ServiceKey, appConfig.Consul.LeaderElectionIntervalInMilliSeconds)
}
