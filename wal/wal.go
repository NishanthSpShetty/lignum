package wal

import (
	"context"
	"encoding/json"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

//contains functions to manage wal logs and storage backend

type Payload struct {
	Topic string
	//should this be in payload
	Id   uint64
	Data string
}

func (p Payload) Json() []byte {
	data, _ := json.Marshal(p)
	return data
}

type Wal struct {
	walQueue <-chan Payload
	walMap   map[string]string
}

//Promote promote current file as the associated topic log
//-- WalPromotionSize promote wal file to backend storage file
func (w *Wal) Promote(topic string) {}

func New(conf config.Wal, queue <-chan Payload) *Wal {
	return &Wal{
		walQueue: queue,
		walMap:   make(map[string]string),
	}
}

func (w *Wal) writeToWal(payload Payload) {
	//read the payload message,
	//get the wal file associated with the topic
	//create if file does not exist.
	//append the message to file
}

func (w *Wal) StartWalWriter(ctx context.Context) {

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("stopping wal writer service")
				return
			case payload := <-w.walQueue:
				w.writeToWal(payload)
			}

		}
	}()
}
