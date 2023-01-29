package wal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

// contains functions to manage wal logs and storage backend

type Payload struct {
	Topic string
	// should this be in payload
	Id      uint64
	Data    string
	Promote bool
}

func (p Payload) Json() []byte {
	data, _ := json.Marshal(p)
	return data
}

type Wal struct {
	dataDir  string
	walQueue <-chan Payload
	walCache *walWriterCache
}

func New(conf config.Wal, logDir string, queue <-chan Payload) *Wal {
	return &Wal{
		dataDir:  logDir,
		walQueue: queue,
		walCache: newCache(),
	}
}

func (w *Wal) createFile(dataDir string, topic string, id uint64) (*os.File, error) {
	path := getTopicDatDir(dataDir, topic)
	err := createPath(path)
	if err != nil {
		return nil, err
	}

	fileOffset := id
	path = fmt.Sprintf("%s/%s_%d.qwal", path, topic, fileOffset)

	file, err := os.Create(path)
	return file, err
}

func (w *Wal) ResetWal(topic string) {
	w.walCache.delete(topic)
}

// getWalWriter return the WAL writer if present.
// create new file and WAL writer for the given topic
func (w *Wal) getWalWriter(payload Payload) *bufio.Writer {
	if wf, ok := w.walCache.get(payload.Topic); ok {
		return wf.writer
	}
	// cache miss, create new WalFile
	f, err := w.createFile(w.dataDir, payload.Topic, payload.Id)
	if err != nil {
		log.Error().Err(err).Str("topic", payload.Topic).Uint64("offset", payload.Id).Msg("failed to create new wal file")
		return nil
	}

	writer := bufio.NewWriter(f)
	w.walCache.set(payload.Topic, &walFile{file: f, writer: writer})
	return writer
}

//Promote promote current file as the associated topic log
//-- WalPromotionSize promote wal file to backend storage file
func (w *Wal) Promote(topic string) error {
	var f *os.File

	if wf, ok := w.walCache.get(topic); ok {
		f = wf.file
		// cleanup immediately as new file would be created by another message if delayed.
		w.ResetWal(topic)
	} else {
		return fmt.Errorf("WAL file does not exist for topic: %s", topic)
	}

	name := f.Name()
	err := os.Rename(name, strings.Replace(name, "qwal", "log", 1))
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) writeToWal(payload Payload) {
	// read the payload message,
	if payload.Promote {
		err := w.Promote(payload.Topic)
		if err != nil {
			log.Error().Err(err).Msg("promote failed")
		}
		return
	}

	// get the wal file associated with the topic
	file := w.getWalWriter(payload)
	// create if file does not exist.
	file.WriteString(fmt.Sprintf("%d%s%s\n", payload.Id, MESSAGE_KEY_VAL_SEPERATOR, payload.Data))
	// append the message to file
	file.Flush()
}

func (w *Wal) StartWalWriter(ctx context.Context) {
	log.Debug().Msg("Started WAL writer service")
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("stopping wal writer service")
				return
			case payload := <-w.walQueue:
				log.Debug().Interface("payload", payload).Msg("write")
				w.writeToWal(payload)
			}
		}
	}()
}
