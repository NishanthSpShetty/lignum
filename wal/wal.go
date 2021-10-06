package wal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

//contains functions to manage wal logs and storage backend

type Payload struct {
	Topic string
	//should this be in payload
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
	walFile  sync.Map
}

type WalFile struct {
	file   *os.File
	writer *bufio.Writer
}

func New(conf config.Wal, logDir string, queue <-chan Payload) *Wal {
	return &Wal{
		dataDir:  logDir,
		walQueue: queue,
		walFile:  sync.Map{},
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

func (w *Wal) getWriter(topic string) *bufio.Writer {
	wf, ok := w.walFile.Load(topic)
	if ok {
		return wf.(*WalFile).writer
	}
	return nil
}

//getWalWriter return the WAL writer if present.
// create new file and WAL writer for the given topic
func (w *Wal) getWalWriter(payload Payload) *bufio.Writer {
	if f, ok := w.walFile.Load(payload.Topic); ok {
		return f.(*WalFile).writer
	}
	f, err := w.createFile(w.dataDir, payload.Topic, payload.Id)
	if err != nil {
		log.Error().Err(err).Str("topic", payload.Topic).Uint64("offset", payload.Id).Msg("failed to create new wal file")
		return nil
	}
	writer := bufio.NewWriter(f)
	w.walFile.Store(payload.Topic, &WalFile{file: f, writer: writer})
	return writer
}

//getWalFile return the *os.File for the topic WAL file
func (w *Wal) getWalFile(topic string) *os.File {
	//TODO: what if does not exist, this should never be called though
	if wf, ok := w.walFile.Load(topic); ok && wf != nil {
		return wf.(*WalFile).file
	}
	return nil
}

func (w *Wal) ResetWal(topic string) {
	w.walFile.Delete(topic)
}

//Promote promote current file as the associated topic log
//-- WalPromotionSize promote wal file to backend storage file
func (w *Wal) Promote(topic string) error {
	f := w.getWalFile(topic)
	if f == nil {
		return fmt.Errorf("WAL file does not exist for topic: %s", topic)
	}
	name := f.Name()
	err := os.Rename(name, strings.Replace(name, "qwal", "log", 1))
	if err != nil {
		return err
	}
	w.ResetWal(topic)
	err = f.Close()
	if err != nil {
		return err
	}
	return nil

}

func (w *Wal) writeToWal(payload Payload) {
	//read the payload message,
	if payload.Promote {
		err := w.Promote(payload.Topic)
		if err != nil {
			log.Error().Err(err).Msg("promote failed")
		}
		return
	}

	//get the wal file associated with the topic
	file := w.getWalWriter(payload)
	//create if file does not exist.
	file.WriteString(fmt.Sprintf("%d%s%s\n", payload.Id, MESSAGE_KEY_VAL_SEPERATOR, payload.Data))
	//append the message to file
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
