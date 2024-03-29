package message

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message/types"
	"github.com/NishanthSpShetty/lignum/metrics"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/rs/zerolog/log"
)

const errBadReplicationStateFmtStr = "bad replication state, expected sequence: %d, got sequence %d"

type MessageStore struct {
	topic             map[string]*types.Topic
	messageBufferSize uint64
	dataDir           string
	walChannel        chan<- wal.Payload
}

// TODO: refactor this
// Payload duplicate payload definition to avoid cyclic import
type Payload struct {
	Topic string
	// should this be in payload
	Id   uint64
	Data []byte
}

func New(msgConfig config.Message, walChannel chan<- wal.Payload) *MessageStore {
	// TODO: restore from the file when we add persistence
	//	messages := ReadFromLogFile(messageConfig.MessageDir)
	return &MessageStore{
		topic:             make(map[string]*types.Topic),
		messageBufferSize: msgConfig.InitialSizePerTopic,
		dataDir:           msgConfig.DataDir,
		walChannel:        walChannel,
	}
}

func (m *MessageStore) BufferSize() uint64 {
	return m.messageBufferSize
}

func getWalFile(topic, path string) (string, uint64) {
	var offset uint64 = 0
	dir, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Msg("getWalFile: failed to open file")
		return "", offset
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return "", offset
	}

	// filter file ending with qwal
	walFilePath := ""
	walFileName := ""
	for _, _file := range files {
		if strings.HasSuffix(_file.Name(), "qwal") {
			walFileName = _file.Name()
			walFilePath = path + "/" + walFileName
		}
	}

	// get the message offset
	offsetStr := strings.ReplaceAll(walFileName, topic+"_", "")
	offsetStr = strings.ReplaceAll(offsetStr, ".qwal", "")
	offset, err = strconv.ParseUint(offsetStr, 10, 64)
	return walFilePath, offset
}

// RestoreWAL on startup read WAL files and replay the messages
// update WalCache accordignly
func (m *MessageStore) RestoreWAL(walP *wal.Wal) {
	// read topics from the data directory
	dadaDir, err := os.Open(m.dataDir)
	if os.IsNotExist(err) {
		// we havent created any data directory, so all good here, return
		return
	}
	dirs, err := dadaDir.Readdir(-1)

	if err != nil && len(dirs) == 0 {
		// if errored or no files returned, just return.
		// it can return err but still return partial list
		return
	}

	for _, topicDir := range dirs {
		if topicDir.IsDir() {
			topicName := topicDir.Name()
			topic := m.createNewTopic(topicName, m.messageBufferSize)
			fmt.Println("Loading topic ", topicName, " from WAL file")
			walFilePath, offset := getWalFile(topicName, m.dataDir+"/"+topicDir.Name())

			log.Debug().Str("file", walFilePath).Str("data_dir", m.dataDir).Msg("found wal file for topic")
			file, err := os.OpenFile(walFilePath, os.O_RDONLY, os.ModePerm)
			if err != nil {
				log.Error().Err(err).Str("filename", walFilePath).Msg("failed to open WAL file")
				continue
			}

			if file == nil {
				// FIXME: too many wal files,
				log.Error().Str("topic", topicName).Msg("wal file not found")
				continue
			}
			// we would have at most messageBufferSize number of messages in WAL file,
			endOffset := offset + m.messageBufferSize
			_ = topic

			raw, err := wal.ReadFromWal(file, offset, endOffset)
			msgs := types.DecodeRawMessage(raw, offset, endOffset)
			if err != nil {
				log.Error().Err(err).Str("file", file.Name()).Msg("error reading wal file")
				continue
			}

			file.Close()
			// load messages back to topic
			topic.PushAll(msgs)
			lastMsg := msgs[len(msgs)-1]
			topic.SetCounter(lastMsg.Id + 1)

			f, err := os.OpenFile(walFilePath, os.O_APPEND|os.O_RDWR, os.ModePerm)
			if err != nil {
				return
			}
			walP.UpdateWalCache(topicName, f)
		}
	}
}

func (m *MessageStore) GetTopics() []*types.Topic {
	topics := make([]*types.Topic, 0)
	for _, v := range m.topic {
		topics = append(topics, v)
	}
	return topics
}

func (m *MessageStore) TopicExist(topic string) bool {
	_, ok := m.topic[topic]
	return ok
}

func (m *MessageStore) createNewTopic(topicName string, msgBufferSize uint64) *types.Topic {
	topic := types.NewTopic(topicName, msgBufferSize, m.dataDir)
	metrics.IncrementTopic()
	m.topic[topicName] = topic
	return topic
}

func (m *MessageStore) CreateNewTopic(topicName string, msgBufferSize uint64, liveReplication bool) {
	topic := m.createNewTopic(topicName, msgBufferSize)
	if liveReplication {
		topic.EnableLiveReplication()
	}
}

// Put write message to store and return the new Message and bool value indicating live replication
func (m *MessageStore) Put(ctx context.Context, topicName string, data []byte) (types.Message, bool) {
	// check if the topic exist
	topic, ok := m.topic[topicName]

	// create new topic if it doesnt exist
	if !ok {
		log.Info().Str("topic", topicName).Msg("topic does not exist, creating")
		topic = m.createNewTopic(topicName, m.messageBufferSize)
	}

	metrics.IncrementMessageCount(topic.GetName())

	topic.Lock()
	defer topic.Unlock()
	currentOffset := topic.GetCurrentOffset()
	if currentOffset != 0 && currentOffset%uint64(topic.GetMessageBufferSize()) == 0 {
		// we have filled the message store buffer, flush to file
		// promote current wal file and reset the buffer
		// signal wal writer to promote current wal file
		m.walChannel <- wal.Payload{
			Topic:   topicName,
			Promote: true,
		}
		topic.ResetMessageBuffer()
	}

	msg := types.Message{Id: topic.CounterNext(), Data: data}
	// push the message onto wal writer queue
	m.walChannel <- wal.Payload{
		Topic: topicName,
		Id:    msg.Id,
		Data:  msg.Data,
	}
	return topic.Push(msg), topic.LiveReplication()
}

// Get return the value for given range (from, to)
// returns value starting with offset `from` to `to` (exclusive)
// Must: from < to
func (m *MessageStore) Get(topicName string, from, to uint64) []*types.Message {
	// 2, 5 => 2,3,4
	topic, ok := m.topic[topicName]

	if !ok {
		return nil
	}

	return topic.GetMessages(from, to)
}

func (m *MessageStore) Replicate(payload Payload) error {
	topic, ok := m.topic[payload.Topic]

	if !ok {
		topic = m.createNewTopic(payload.Topic, m.messageBufferSize)
	}

	// assert that we got expected message sequence.
	if topic.GetCurrentOffset() != payload.Id {
		return fmt.Errorf(errBadReplicationStateFmtStr, topic.GetCurrentOffset(), payload.Id)
	}

	// metrics.IncrementMessageCount(t.name)
	message := types.Message{Id: topic.CounterNext(), Data: payload.Data}
	topic.Append(message)
	return nil
}

func (m *MessageStore) WalMetaUpdate(topicName string, offset uint64) {
	topic, ok := m.topic[topicName]

	if !ok {
		topic = m.createNewTopic(topicName, m.messageBufferSize)
	}
	topic.SetCounter(offset)
}
