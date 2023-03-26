package replication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/rs/zerolog/log"
)

type Payload struct {
	Topic string
	// should this be in payload
	Id   uint64
	Data []byte
}

func (p Payload) Json() []byte {
	data, _ := json.Marshal(p)
	return data
}

// ReplicationState Contains the information on the followers replciation state
type ReplicationState struct {
	node types.Node
	// mark that replicator can start the replication for this node
	ready bool
	// message offset which is already been sent to follower
	offset int64
}

type LiveReplicator struct {
	replicationQueue <-chan Payload
	//	replicationState map[string]ReplicationState
	followerRegistry *follower.FollowerRegistry
	client           http.Client
}

// Create a new live replication which implements active replication of a topic queue
func NewLiveReplicator(queue <-chan Payload, followerRegistry *follower.FollowerRegistry) *LiveReplicator {
	return &LiveReplicator{replicationQueue: queue, followerRegistry: followerRegistry}
}

// Start start replication routine to replicate the messages to all nodes
func (r *LiveReplicator) Start(ctx context.Context, replicationTimeoutInMs time.Duration) {
	r.client = http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: replicationTimeoutInMs,
	}

	log.Info().Msg("replicator service is running..")
	// this is a live replication, so when the leader recieves the message it will write the same to the follower.
	// This doesnt care for what the replication state of the follower
	// follower must catch up with leader if there is a lag
	go func() {
		// when queue is closed, the following looping would stop, effectively stopping the routine
		for payload := range r.replicationQueue {
			log.Debug().Interface("payload", payload).Msg("received message for replication ")
			r.replicate(payload)
		}
	}()
}

func (r *LiveReplicator) replicate(payload Payload) {
	for id, follower := range r.followerRegistry.List() {
		log.Debug().Str("follower_service_id", id).Msg("sending message to follower")
		if follower.IsReady() {
			r.send(follower.Node(), payload)
		}
	}
}

func (r *LiveReplicator) send(node types.Node, payload Payload) {
	url := fmt.Sprintf("http://%s:%d/internal/api/replicate", node.Host, node.Port)
	contentType := "application/json"

	response, err := r.client.Post(url, contentType, bytes.NewBuffer(payload.Json()))
	if err != nil {
		log.Error().RawJSON("node", node.Json()).Err(err).Msg("failed to send message to follower")
		// should add some retrier mechanism
		return
	}
	b, _ := ioutil.ReadAll(response.Body)
	reason := string(b)
	if response.StatusCode == http.StatusBadRequest {
		log.Error().Str("reason", reason).RawJSON("node", node.Json()).Msg("follower rejected replication message")
	}
	if response.StatusCode == http.StatusOK {
		log.Debug().Msg("msg sent to follower")
	}
	response.Body.Close()
}
