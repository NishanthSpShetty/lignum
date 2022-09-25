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

	buf := make([]byte, 128)

	//indicate when to clean storage buf
	clean := false
	for {
		meta := make([]byte, 1024)
		n, err := c.Read(meta)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error().Err(err).Msg("connection closed.")
				return
			}
			log.Error().Err(err).Msg("failed to read the wal metadata")
			continue
		}
		fmt.Printf("read %d bytes \n", n)

		marker := false
		for _, b := range meta {

			marker = false
			if isMarker(b) {
				marker = true
			}

			//if the current one is end marker, check prev bytes
			if marker {
				l := len(buf)
				if b == MARKER_META_END {
					fmt.Println("meta end marker detected")
					if buf[1-1] == MARKER2 && buf[l-2] == MARKER1 {
						//delimiter found, grab a chunk and process it
						md, err := ToMeta(buf[0 : l-2])
						if err != nil {
							log.Error().Err(err).Msg("failed to parse meta")
							continue
						}
						fmt.Printf("Wal meta %v \n", md)
						clean = true
					}
				} else if b == MARKER_FILE_END {
					fmt.Println("file end marker detected")
					if buf[1-1] == MARKER2 && buf[l-2] == MARKER1 {
						//delimiter found, grab a chunk and process it
						fileData := buf[0 : l-2]
						fmt.Println("content of wal ", string(fileData))
					}
					clean = true
				}

			}
			if clean {
				buf = make([]byte, 128)
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

		if err != nil {
			log.Error().Err(err).Msg("failed to accept connection")
			continue
		}

		handleClient(client)

	}

}

//wal replication service
func (w *WalService) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", w.conf.Server.Host, w.conf.Replication.WALReplicationPort)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return errors.Wrap(err, "Start")
	}

	go start(ctx, listener)
	return nil
}
