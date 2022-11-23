package replication

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var OffsetExtractPattern *regexp.Regexp

// responsibility
// create a wal file replicator routine
// it will
// 1. Sync up the topic queue to each follower post registration
// 2. Send the newly created wal file on wal file creation on new.

type WALReplicator struct {
	fq                    chan *follower.Follower
	client                http.Client
	syncFollwers          []*follower.Follower
	failedFollwers        []*follower.Follower
	signalLeader          chan bool
	lock                  sync.RWMutex
	syncIntervalInSeconds time.Duration
}

func NewWALReplication(fq chan *follower.Follower, leaderSignal chan bool, syncIntervalInSeconds time.Duration) *WALReplicator {
	// need a queue on which we get newly registered follower
	OffsetExtractPattern = regexp.MustCompile(`(\d+)`)
	return &WALReplicator{
		fq:                    fq,
		signalLeader:          leaderSignal,
		lock:                  sync.RWMutex{},
		syncIntervalInSeconds: syncIntervalInSeconds,
	}
}

func getOffsetFromFile(name string) uint64 {
	foundStr := OffsetExtractPattern.Find([]byte(name))
	offset, err := strconv.ParseUint(string(foundStr), 10, 64)
	if err != nil {
		return 0
	}
	return offset
}

func (w *WALReplicator) failedFollwer(f *follower.Follower) {
	w.failedFollwers = append(w.failedFollwers, f)
}

func (w *WALReplicator) syncedFollwer(f *follower.Follower) {
	w.lock.Lock()
	defer w.lock.Unlock()
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

func (w *WALReplicator) syncFollwer(msgStore *message.MessageStore, f *follower.Follower, synced bool) {
	// get the current topic in system
	//possible that the topics have updated after last follower registred, so we need to query the new state of the system as soon as new follower registers
	currentTopics := msgStore.GetTopics()
	log.Info().
		Int("topics", len(currentTopics)).
		Str("", f.Node().Host).
		Str("id", f.Node().Id).
		Msg("syncing the follower")

	//get the tcp connection
	addr := fmt.Sprintf("%s:%d", f.Node().Host, f.Node().ReplicationPort)
	//TODO: we can have a iteration where no topic data is there to replicate in such case obtaining connection and doing other work is waste, optimise it.
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
		//check if files returned is empty, if the stat is upto date with the system we wont get any wal files

		lastWalOffset := uint64(0)

		for _, file := range files {
			fileName := filepath.Base(file)
			fmt.Println("sending file ", fileName)
			log.Debug().Str("filename", fileName).Str("path", file).Msg("sending wal file")

			//get the topic wal files.
			fil, err := os.Open(file)
			if err != nil {
				log.Error().Err(err).Msg("failed to open file")
				break

			}

			lastWalOffset = getOffsetFromFile(fileName) + topic.GetMessageBufferSize()

			meta := wal.Metadata{
				Topic:      topic.GetName(),
				WalFile:    fileName,
				NextOffSet: lastWalOffset,
			}

			metad, err := meta.Bytes()
			if err != nil {
				log.Error().Err(err).Msg("failed to encode metdata into bytes")
			}
			conn.Write(metad)
			conn.Write(wal.MetaMarker())

			fi, err := fil.Stat()
			if err != nil {
				log.Error().Err(err).Msg("failed to get the file stat")
				continue
			}
			err = sendFile(conn.(*net.TCPConn), fil, fi)
			if err != nil {
				// FIXME : what we do here ?
				log.Error().Err(err).
					Str("topic", topic.GetName()).
					Str("file", file).
					Msg("failed to send wal file ")
			}
			conn.Write(wal.FileMarker())
			fil.Close()
		}

		//get the topic last message offset of synced wal
		log.Info().
			Str("topic", topic.GetName()).
			Str("id", f.Node().Id).
			Msg("synced topic")

		if lastWalOffset != 0 {
			log.Debug().
				Int("offset", int(lastWalOffset)).
				Str("topic", topic.GetName()).
				Str("node", f.Node().Id).
				Msg("updating message stat for node")
			f.UpdateStat(topic.GetName(), lastWalOffset)
		}
	}

	if !synced {
		//add the updated follower to synced followers list
		w.syncedFollwer(f)
		//mark the node as ready
		f.MarkReady()
	}

	conn.Close()
}

func (w *WALReplicator) topicSyncerForNewFollower(msgStore *message.MessageStore) {
	//sync new follower when registers
	j := 0
	for f := range w.fq {
		fmt.Println("Syncing new follower in iteration ", "count ", j)
		w.syncFollwer(msgStore, f, false)
		j++
	}
}

func (w *WALReplicator) topicSyncer(msgStore *message.MessageStore) {
	iteration := 0
	for {
		//loop over synced topic and replicate any new wal file present
		// FIX: potential race, writes happening in other thread
		if len(w.syncFollwers) == 0 {
			log.Debug().Msg("there are no follower to sync, going to sleep")
			time.Sleep(time.Second * 3)
			continue
		}

		// as the below operation takes time, we can let the writer update the slice by copying and using the copy
		followers := make([]*follower.Follower, len(w.syncFollwers))
		copy(followers, w.syncFollwers)
		for _, f := range followers {
			if f.IsReady() && f.IsHealthy() {
				w.syncFollwer(msgStore, f, true)
			}
			if !f.IsHealthy() {
				log.Debug().Str("node", f.Node().Id).
					Msg("unhealthy node in follower list, skipped")
			}
		}
		log.Debug().
			Int("iteration", iteration).
			Int("followers", len(followers)).
			Msg("iteration sync complete")
		iteration++
		time.Sleep(time.Second * w.syncIntervalInSeconds)
	}
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

		go w.topicSyncerForNewFollower(msgStore)
		w.topicSyncer(msgStore)
	}()
	return nil
}
