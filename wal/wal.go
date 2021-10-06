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
	dataDir   string
	walQueue  <-chan Payload
	walWriter map[string]*bufio.Writer
	walFile   map[string]*os.File
}

func New(conf config.Wal, logDir string, queue <-chan Payload) *Wal {
	return &Wal{
		dataDir:   logDir,
		walQueue:  queue,
		walWriter: make(map[string]*bufio.Writer),
		walFile:   make(map[string]*os.File),
	}
}

func (w *Wal) createFile(dataDir string, topic string, id uint64) (*os.File, error) {

	path := getTopicDatDir(dataDir, topic)
	err := createPath(path)

	if err != nil {
		return nil, err
	}

	fileOffset := id
	path = fmt.Sprintf("%s/%s_%d.log", path, topic, fileOffset)

	file, err := os.Create(path)
	return file, err
}

//getWalWriter return the WAL writer if present.
// create new file and WAL writer for the given topic
func (w *Wal) getWalWriter(payload Payload) *bufio.Writer {
	if f, ok := w.walWriter[payload.Topic]; ok {
		return f
	}
	f, err := w.createFile(w.dataDir, payload.Topic, payload.Id)
	if err != nil {
		log.Err(err).Str("topic", payload.Topic).Uint64("offset", payload.Id).Msg("failed to create new wal file")
		return nil
	}
	writer := bufio.NewWriter(f)
	w.walWriter[payload.Topic] = writer
	w.walFile[payload.Topic] = f
	return writer
}

//getWalFile return the *os.File for the topic WAL file
func (w *Wal) getWalFile(topic string) *os.File {
	//TODO: what if does not exist, this should never be called though
	return w.walFile[topic]
}

//Promote promote current file as the associated topic log
//-- WalPromotionSize promote wal file to backend storage file
func (w *Wal) Promote(topic string) {
	f := w.getWalFile(topic)
	name := f.Name()
	os.Rename(name, strings.Replace(name, "wal", "log", 1))
	f.Close()

}

func (w *Wal) writeToWal(payload Payload) {
	//read the payload message,
	if payload.Promote {
		w.Promote(payload.Topic)
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
				w.writeToWal(payload)
			}

		}
	}()
}
