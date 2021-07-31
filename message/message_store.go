package message

import (
	"context"
	"fmt"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message/types"
	"github.com/NishanthSpShetty/lignum/metrics"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/rs/zerolog/log"
)

const errBadReplicationStateFmtStr = "bad replication state, expected sequence: %d, got sequence %d"

type MessageStore struct {
	topic             map[string]*Topic
	messageBufferSize uint64
	dataDir           string
}

func New(msgConfig config.Message) *MessageStore {
	//TODO: restore from the file when we add persistence
	//	messages := ReadFromLogFile(messageConfig.MessageDir)
	return &MessageStore{
		topic:             make(map[string]*Topic),
		messageBufferSize: msgConfig.InitialSizePerTopic,
		dataDir:           msgConfig.DataDir,
	}
}

func (m *MessageStore) GetTopics() []*Topic {
	topics := make([]*Topic, 0)
	for _, v := range m.topic {
		topics = append(topics, v)
	}
	return topics
}

func (m *MessageStore) TopicExist(topic string) bool {
	_, ok := m.topic[topic]
	return ok
}

func (m *MessageStore) createNewTopic(topicName string, msgBufferSize uint64) *Topic {

	topic := &Topic{
		name:          topicName,
		counter:       NewCounter(),
		messageBuffer: make([]types.Message, msgBufferSize),
		msgBufferSize: msgBufferSize,
		dataDir:       m.dataDir,
	}
	metrics.IncrementTopic()
	m.topic[topicName] = topic
	return topic
}

func (m *MessageStore) Put(ctx context.Context, topicName string, msg string) types.Message {
	//check if the topic exist
	topic, ok := m.topic[topicName]

	//create new topic if it doesnt exist
	if !ok {
		log.Info().Str("Topic", topicName).Msg("topic does not exist, creating")
		topic = m.createNewTopic(topicName, m.messageBufferSize)
	}

	metrics.IncrementMessageCount(topic.name)

	if topic.counter.value%uint64(topic.msgBufferSize) == 0 {
		// we have filled the message store buffer, flush to file
		msgBuffer := topic.messageBuffer
		topic.resetMessageBuffer()
		writeToLogFile(m.dataDir, topic.name, msgBuffer)
	}
	return topic.Push(msg)
}

//Get return the value for given range (from, to)
//returns value starting with offset `from` to `to` (exclusive)
//Must: from < to
func (m *MessageStore) Get(topicName string, from, to uint64) []*types.Message {
	// 2, 5 => 2,3,5
	topic, ok := m.topic[topicName]

	if !ok {
		return nil
	}

	return topic.GetMessages(from, to)
}

func (m *MessageStore) Replicate(payload replication.Payload) error {
	topic, ok := m.topic[payload.Topic]

	if !ok {
		topic = m.createNewTopic(payload.Topic, m.messageBufferSize)
	}

	//assert that we got expected message sequence.
	if topic.counter.value != payload.Id {
		return fmt.Errorf(errBadReplicationStateFmtStr, topic.counter.value, payload.Id)
	}

	//metrics.IncrementMessageCount(t.name)
	message := types.Message{Id: topic.counter.Next(), Data: payload.Data}
	topic.messageBuffer = append(topic.messageBuffer, message)
	return nil
}
