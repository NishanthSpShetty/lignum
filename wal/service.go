package wal

import (
	"context"
	"fmt"
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

func handleClient(c net.Conn) {

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
