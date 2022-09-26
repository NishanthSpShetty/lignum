package replication

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	signalLeader   chan bool
}

func NewWALReplication(fq chan *follower.Follower, leaderSignal chan bool) *WALReplicator {
	// need a queue on which we get newly registered follower
	return &WALReplicator{
		fq:           fq,
		signalLeader: leaderSignal,
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
	for {
		for f := range w.fq {
			if f.IsReady() {
				fmt.Println("follower is ready, skipping", "id", f.Node().Id)
				continue
			}
			// get the current topic in system
			//possible that the topics have updated after last follower registred, so we need to query the new state of teh system as soon as new follower registers
			currentTopics := msgStore.GetTopics()
			log.Info().
				Int("topics", len(currentTopics)).
				Str("", f.Node().Host).
				Str("id", f.Node().Id).
				Msg("syncing the follower")

			//get the tcp connection
			addr := fmt.Sprintf("%s:%d", f.Node().Host, f.Node().ReplicationPort)
			conn, err := net.Dial("tcp", addr)

			if err != nil {
				log.Error().Err(err).
					Str("id", f.Node().Id).
					Msg("unable to create tcp connection with the follower")
				w.failedFollwer(f)
			}
			log.Info().Msg("connection established with follower")

			//reconcile the topics
			for _, topic := range currentTopics {
				log.Debug().Str("topic", topic.GetName()).Msg("syncing topic to follower")

				//should be the last message offset in the latest wal file
				currentOffset := f.TopicOffset(topic.GetName())
				files := topic.GetWalFile(currentOffset)

				for _, file := range files {
					fileName := filepath.Base(file)
					fmt.Println("sending file ", fileName)
					log.Debug().Str("filename", fileName).Str("path", file).Msg("sending wal file")

					//get the topic wal files.
					f, err := os.Open(file)
					if err != nil {
						log.Error().Err(err).Msg("failed to open file")
						break
					}
					meta := wal.Metadata{
						Topic:   topic.GetName(),
						WalFile: fileName,
					}

					metad, err := meta.Bytes()
					if err != nil {
						log.Error().Err(err).Msg("failed to encode metdata into bytes")
					}
					conn.Write(metad)
					conn.Write(wal.MetaMarker())

					fi, err := f.Stat()
					if err != nil {
						log.Error().Err(err).Msg("failed to get the file stat")
						continue
					}
					err = sendFile(conn.(*net.TCPConn), f, fi)
					if err != nil {
						// FIXME : what we do here ?
						log.Error().Err(err).
							Str("topic", topic.GetName()).
							Str("file", file).
							Msg("failed to send wal file ")
					}
					//write end of file marker
					conn.Write(wal.FileMarker())
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
			conn.Close()
		}
	}
	//for every topic, send all wal files to follower node
}

//Start starting the wal replicator service in leader only
func (w *WALReplicator) Start(ctx context.Context, replicationTimeoutInMs time.Duration, msgStore *message.MessageStore) error {
	w.client = http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: replicationTimeoutInMs,
	}

	//wait till this node becomes a leader
	go func() {
		<-w.signalLeader
		log.Info().Msg("starting wal replicator routine in leader")

		w.topicSyncer(msgStore)
	}()
	return nil
}
