package wal

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// all follower must start service to accept wal files from the leader

type WalService struct {
	conf config.Config
}

func NewReplicaionService(c config.Config) *WalService {
	return &WalService{conf: c}
}

//handleClient for the connected client read all incoming wal request
func handleClient(c net.Conn) {

	//read the meta data
	//read the file associated with it,
	// format
	// metadata + file content

	buf := make([]byte, 0)
	//indicate when to clean storage buf
	clean := false
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

			marker = isMarker(b)

			//if the current one is end marker, check prev bytes
			if marker {
				l := len(buf)
				if b == MARKER_META_END {
					if buf[l-1] == MARKER2 && buf[l-2] == MARKER1 {
						//delimiter found, grab a chunk and process it
						m := make([]byte, l-2)
						copy(m, buf[0:l-2])
						md, err := ToMeta(m)
						if err != nil {
							log.Error().Err(err).Msg("failed to parse meta")
							continue
						}
						fmt.Printf("\nWal meta %+v \n", md)
						clean = true
					}
				}
				if b == MARKER_FILE_END {
					if buf[l-1] == MARKER2 && buf[l-2] == MARKER1 {
						//delimiter found, grab a chunk and process it
						fileData := buf[0 : l-2]
						fmt.Println("Content of wal\n", string(fileData))
						clean = true
					}
				}

			}
			if clean {
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

func start(ctx context.Context, listener net.Listener) {

	for {
		client, err := listener.Accept()
		log.Info().
			Str("client", client.LocalAddr().String()).
			Msg("accepted connection")

		if err != nil {
			log.Error().Err(err).Msg("failed to accept connection")
			continue
		}

		handleClient(client)

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
	go start(ctx, listener)
	return nil
}
