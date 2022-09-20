package replication

import (
	"context"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/rs/zerolog/log"
)

// responsibility
// create a wal file replicator routine
// it will
// 1. Sync up the topic queue to each follower post registration
// 2. Send the newly created wal file on wal file creation on new.

type WALReplicator struct {
	fq     chan *follower.Follower
	client http.Client
}

func NewWALReplication(fq chan *follower.Follower) *WALReplicator {
	// need a queue on which we get newly registered follower
	return &WALReplicator{
		fq: fq,
	}
}

func (w *WALReplicator) Start(_ context.Context, replicationTimeoutInMs time.Duration) error {
	w.client = http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: replicationTimeoutInMs,
	}

	go func() {
		log.Info().Msg("starting wal replicator routine")
	}()
	return nil
}
