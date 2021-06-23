package message

import (
	"context"
	"fmt"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/metrics"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/rs/zerolog/log"
)

const errBadReplicationStateFmtStr = "bad replication state, expected sequence: %d, got sequence %d"

type Message struct {
	Id uint64
	//TODO: consider []byte here
	Data string
}

func (m Message) String() string {
	return fmt.Sprintf("{ID: %v, Msg: %s}\n", m.Id, m.Data)
}

type Topic struct {
	counter *Counter
	name    string
	msg     []Message
}

func (t *Topic) GetName() string {
	return t.name
}

func (t *Topic) GetCurrentOffset() uint64 {
	return t.counter.value
}

func (t *Topic) GetMessages() []Message {
	return t.msg
}

func (t *Topic) Push(msg string) Message {
	metrics.IncrementMessageCount(t.name)
	message := Message{Id: t.counter.Next(), Data: msg}
	t.msg = append(t.msg, message)
	return message
}

type MessageStore struct {
	topic map[string]*Topic
}

func New(msgConfig config.Message) *MessageStore {
	//TODO: restore from the file when we add persistence
	//	messages := ReadFromLogFile(messageConfig.MessageDir)
	return &MessageStore{
		topic: make(map[string]*Topic),
	}
}

func (m *MessageStore) GetTopics() []*Topic {
	topics := make([]*Topic, 0)
	for _, v := range m.topic {
		topics = append(topics, v)
	}
	return topics
}

func (m *MessageStore) GetMessages(topicName string) []Message {

	topic, ok := m.topic[topicName]
	if !ok {
		return []Message{}
	}
	return topic.GetMessages()
}

func (m *MessageStore) TopicExist(topic string) bool {
	_, ok := m.topic[topic]
	return ok
}

func (m *MessageStore) createNewTopic(topic_name string) *Topic {

	topic := &Topic{
		name:    topic_name,
		counter: NewCounter(),
		msg:     make([]Message, 0),
	}
	metrics.IncrementTopic()
	m.topic[topic_name] = topic
	return topic
}

func (m *MessageStore) Put(ctx context.Context, topic_name string, msg string) Message {
	//check if the topic exist
	topic, ok := m.topic[topic_name]

	//create new topic if it doesnt exist
	if !ok {
		log.Info().Str("Topic", topic_name).Msg("topic does not exist, creating")
		topic = m.createNewTopic(topic_name)
	}

	//push message into topic
	return topic.Push(msg)
}

//Get return the value for given range (from, to)
//returns value starting with offset `from` to `to` (exclusive)
//Must: from < to
func (m *MessageStore) Get(topic string, from, to uint64) []Message {
	// 2, 5 => 2,3,5
	messages := m.GetMessages(topic)
	msgLen := uint64(len(messages))

	if msgLen == 0 {
		return []Message{}
	}
	if to > msgLen {
		to = msgLen
	}
	msgs := make([]Message, to-from)
	i := 0
	for _, msg := range messages {
		if msg.Id < from {
			continue
		}
		if msg.Id >= to {
			break
		}
		msgs[i] = msg
		i++
	}
	return msgs
}

func (m *MessageStore) Replicate(payload replication.Payload) error {
	topic, ok := m.topic[payload.Topic]

	if !ok {
		topic = m.createNewTopic(payload.Topic)
	}

	//assert that we got expected message sequence.
	if topic.counter.value != payload.Id {
		return fmt.Errorf(errBadReplicationStateFmtStr, topic.counter.value, payload.Id)
	}

	//metrics.IncrementMessageCount(t.name)
	message := Message{Id: topic.counter.Next(), Data: payload.Data}
	topic.msg = append(topic.msg, message)
	return nil
}

//TODO: move out of here
//StartFlusher start flusher routine to write the messages to file
func StartFlusher(messageConfig config.Message) {

	//	go func(messageConfig config.Message) {
	//		for {
	//			time.Sleep(messageConfig.MessageFlushIntervalInMilliSeconds * time.Millisecond)
	//
	//			//keep looping on the above sleep interval when the message size is zero
	//			if len(m.messages) == 0 {
	//				continue
	//			}
	//
	//			count, err := WriteToLogFile(messageConfig, m.messages)
	//			if err != nil {
	//				log.Error().Err(err).Msg("failed to write the messages to file")
	//				continue
	//			}
	//			log.Debug().Int("Count", count).Msg("Wrote %d messages to file")
	//
	//		}
	//	}(messageConfig)
}
