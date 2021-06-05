package message

import (
	"context"
	"fmt"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

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

func (t *Topic) GetMessages() []Message {
	return t.msg
}

func (t *Topic) Push(msg string) {
	message := Message{Id: t.counter.Next(), Data: msg}
	t.msg = append(t.msg, message)
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

func (m *MessageStore) Put(ctx context.Context, topic_name string, msg string) {
	//check if the topic exist
	topic, ok := m.topic[topic_name]

	//create new topic if it doesnt exist
	if !ok {
		log.Info().Str("Topic", topic_name).Msg("topic does not exist, creating")
		topic = &Topic{
			name:    topic_name,
			counter: NewCounter(),
			msg:     make([]Message, 0),
		}
		m.topic[topic_name] = topic
	}

	//push message into topic
	topic.Push(msg)
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
