package replication

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// responsibility
// create a wal file replicator routine
// it will
// 1. Sync up the topic queue to each follower post registration
// 2. Send the newly created wal file on wal file creation on new.

type WALReplicator struct {
	fq             chan *follower.Follower
	client         http.Client
	syncFollwers   []*follower.Follower
	failedFollwers []*follower.Follower
}

func NewWALReplication(fq chan *follower.Follower) *WALReplicator {
	// need a queue on which we get newly registered follower
	return &WALReplicator{
		fq: fq,
	}
}

func (w *WALReplicator) failedFollwer(f *follower.Follower) {
	w.failedFollwers = append(w.failedFollwers, f)
}

func (w *WALReplicator) syncedFollwer(f *follower.Follower) {
	w.syncFollwers = append(w.syncFollwers, f)
}

func sendFile(c *net.TCPConn, f *os.File, fi os.FileInfo) error {
	sockFile, err := c.File()
	if err != nil {
		errors.Wrap(err, "could not get connection file")
	}
	defer sockFile.Close()
	_, err = syscall.Sendfile(int(sockFile.Fd()), int(f.Fd()), nil, int(fi.Size()))

	return errors.Wrap(err, "sendFile")
}

func (w *WALReplicator) topicSyncer(msgStore *message.MessageStore) {
	//sync new follower when registers
	for f := range w.fq {
		// get the current topic in system
		//possible that the topics have updated after last follower registred, so we need to query the new state of teh system as soon as new follower registers
		currentTopics := msgStore.GetTopics()
		log.Info().
			Int("topics", len(currentTopics)).
			Str("", f.Node().Host).
			Str("id", f.Node().Id).
			Msg("syncing the follower")

		//get the tcp connection
		addr := fmt.Sprintf("%s:%d", f.Node().Host, f.Node().Port)
		conn, err := net.Dial("tcp", addr)

		if err != nil {
			log.Error().Err(err).
				Str("id", f.Node().Id).
				Msg("unable to create tcp connection with the follower")
			w.failedFollwer(f)
		}

		//reconcile the topics
		for _, topic := range currentTopics {

			//should be the last message offset in the latest wal file
			currentOffset := f.TopicOffset(topic.GetName())
			//fetch wall files after currentOffset
			files := topic.GetWalFile(currentOffset + 1)

			for _, file := range files {
				//get the topic wal files.
				f, err := os.Open(file)
				if err != nil {
					return
				}
				meta := wal.Metadata{
					Topic:   topic.GetName(),
					WalFile: file,
				}

				metad, err := meta.Bytes()
				if err != nil {
					log.Error().Err(err).Msg("failed to encode metdata into bytes")
				}
				conn.Write(metad)
				conn.Write(wal.Marker())

				fi, err := f.Stat()
				err = sendFile(conn.(*net.TCPConn), f, fi)
				if err != nil {
					// FIXME : what we do here ?
					log.Error().Err(err).
						Str("topic", topic.GetName()).
						Str("file", file).
						Msg("failed to send wal file ")
				}
				//write end of file marker

				conn.Write(wal.Marker())

			}
			log.Info().
				Str("topic", topic.GetName()).
				Str("id", f.Node().Id).
				Msg("synced topic")
		}

		//add the updated follower to synced followers list
		w.syncedFollwer(f)
		//mark the node as ready
		f.MarkReady()
	}
	//for every topic, send all wal files to follower node
}

//Start starting the wal replicator service
func (w *WALReplicator) Start(ctx context.Context, replicationTimeoutInMs time.Duration, msgStore *message.MessageStore) error {
	w.client = http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: replicationTimeoutInMs,
	}

	go func() {
		log.Info().Msg("starting wal replicator routine")

		go w.topicSyncer(msgStore)
		for {

			//need topics, wal files for the topic, current stat of the follower node
		}
	}()
	return nil
}
