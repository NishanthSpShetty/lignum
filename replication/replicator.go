package replication

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/rs/zerolog/log"
)

type Payload struct {
	Topic string
	//should this be in payload
	Id   uint64
	Data string
}

//ReplicationState Contains the information on the followers replciation state
type ReplicationState struct {
	node cluster.Node
	//mark that replicator can start the replication for this node
	ready bool
	//message offset which is already been sent to follower
	offset int64
}

type Replicator struct {
	replicationQueue <-chan Payload
	//	replicationState map[string]ReplicationState
	followerRegistry *follower.FollowerRegistry
	client           http.Client
}

func New(queue <-chan Payload, followerRegistry *follower.FollowerRegistry) *Replicator {
	return &Replicator{replicationQueue: queue, followerRegistry: followerRegistry}
}

//StartReplicator start replication routine to replicate the messages to all nodes
func (r *Replicator) StartReplicator(ctx context.Context, replicationTimeoutInMs time.Duration) {

	r.client = http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: replicationTimeoutInMs,
	}

	log.Info().Msg("replicator service is running..")
	//this is a live replication, so when the leader recieves the message it will write the same to the follower.
	//This doesnt care for what the replication state of the follower
	//follower must catch up with leader if there is a lag
	go func() {
		for payload := range r.replicationQueue {
			log.Debug().Interface("Payload", payload).Msg("received message for replication ")
			r.replicate(payload)
		}
	}()
}

func (r *Replicator) replicate(payload Payload) {
	for id, follower := range r.followerRegistry.List() {
		log.Debug().Str("FollowerServiceID", id).Msg("sending message to follower")
		if follower.IsReady() {
			r.send(follower.Node(), payload)
		}
	}
}

func (r *Replicator) send(node cluster.Node, msg Payload) {

	url := fmt.Sprintf("http://%s:%d/internal/api/replicate", node.Host, node.Port)
	contentType := "application/json"

	response, err := r.client.Post(url, contentType, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Error().RawJSON("node", node.Json()).Err(err).Msg("failed to send message to follower")
		//should add some retrier
	}

	if response.StatusCode == http.StatusOK {
		log.Debug().Msg("msg sent to follower")
	}
}
