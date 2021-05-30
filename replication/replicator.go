package replication

import (
	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/rs/zerolog/log"
)

//ReplicationState Contains the information on the followers replciation state
type ReplicationState struct {
	node cluster.Node
	//mark that replicator can start the replication for this node
	ready bool
	//message offset which is already been sent to follower
	offset int64
}

type Replicator struct {
	replicationQueue <-chan message.Message
	replicationState map[string]ReplicationState
}

func New(queue <-chan message.Message) *Replicator {
	return &Replicator{replicationQueue: queue}
}

//StartReplicator start replication routine to replicate the messages to all nodes
func (r *Replicator) StartReplicator() {

	log.Info().Msg("Replicator service is running..")
	go func() {
		for msg := range r.replicationQueue {
			log.Debug().Interface("Message", msg).Msg("Recieved message for replication ")
		}
	}()
}
