package service

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// all follower must start service to accept wal files from the leader

type WalService struct {
	conf   config.Config
	mstore *message.MessageStore
}

func NewReplicaionService(c config.Config, mstore *message.MessageStore) *WalService {
	return &WalService{conf: c,
		mstore: mstore,
	}
}

func (w *WalService) updateMessageStore(md wal.Metadata) {
	w.mstore.WalMetaUpdate(md.Topic, md.NextOffSet)
}

//handleClient for the connected client read all incoming wal request
// We will recieve multiple wal files of different topics, need to proces everything here
// tried to move the task to wal package however dealing with messageStore will become painful due to the cyclic deps.

func (w *WalService) handleClient(c net.Conn) {

	//read the meta data
	//read the file associated with it,
	// format
	// metadata + file content

	buf := make([]byte, 0)
	//indicate when to clean storage buf
	clean := false
	var data []byte
	var md wal.Metadata
	for {
		meta := make([]byte, 1024)
		_, err := c.Read(meta)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error().Err(err).Msg("connection closed.")
				return
			}
			log.Error().Err(err).Msg("failed to read the wal metadata")
			continue
		}

		marker := false
		for _, b := range meta {

			marker = wal.IsMarker(b)

			//if the current one is end marker, check prev bytes
			if marker {
				l := len(buf)
				if b == wal.MARKER_META_END {
					if buf[l-1] == wal.MARKER2 && buf[l-2] == wal.MARKER1 {
						//delimiter found, grab a chunk and process it
						m := make([]byte, l-2)
						copy(m, buf[0:l-2])
						md, err = wal.ToMeta(m)
						if err != nil {
							log.Error().Err(err).Msg("failed to parse meta")
							continue
						}
						fmt.Printf("\nWal meta %+v \n", md)
						clean = true
					}
				}
				if b == wal.MARKER_FILE_END {
					if buf[l-1] == wal.MARKER2 && buf[l-2] == wal.MARKER1 {
						//delimiter found, grab a chunk and process it
						data = buf[0 : l-2]
						//fmt.Println("Content of wal\n", string(data))
						clean = true
					}
				}

			}
			if clean {
				w.updateMessageStore(md)
				err = wal.WriteWal(w.conf.Message.DataDir, md, data)
				if err != nil {
					//FIXME :we need to do something when this happens ?
					log.Error().Err(err).Msg("failed to write wal to disk")
				}

				//update message store with the updated info

				buf = make([]byte, 0)
				clean = false
			} else {
				buf = append(buf, b)
			}
		}

		if err != nil {
			log.Error().Err(err).Msg("failed to get the wal file attached")
			continue
		}
	}
}

func (w *WalService) start(ctx context.Context, listener net.Listener) {

	for {
		client, err := listener.Accept()
		log.Info().
			Str("client", client.LocalAddr().String()).
			Msg("accepted connection")

		if err != nil {
			log.Error().Err(err).Msg("failed to accept connection")
			continue
		}

		w.handleClient(client)

	}

}

//wal replication service
// Start follower must start this routine to setup wal replication from the leader
func (w *WalService) Start(ctx context.Context) error {

	addr := fmt.Sprintf("%s:%d", w.conf.Server.Host, w.conf.Replication.WALReplicationPort)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Error().Err(err).Send()
		return errors.Wrap(err, "Start")
	}
	//FIXME: this must be killed as soon as the node is promoted to leader
	log.Info().Msg("starting WalService")
	go w.start(ctx, listener)
	return nil
}
