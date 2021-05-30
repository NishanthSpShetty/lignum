package message

import (
	"context"
	"fmt"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

var counter *Counter

type Message struct {
	Id      uint64
	Message string
}

func (m Message) String() string {
	return fmt.Sprintf("{ID: %v, Msg: %s}\n", m.Id, m.Message)
}

var messages []Message

func Init(messageConfig config.Message) {
	log.Info().
		Interface("MessageConfig", messageConfig).
		Msg("Initializing messages")

	//	message = make(MessageT)
	messages = ReadFromLogFile(messageConfig.MessageDir)
	count := len(messages)
	counter = NewCounterWithValue(uint64(count))

}

func Put(ctx context.Context, msg string) {
	messages = append(messages, Message{counter.Next(), msg})
	//	fmt.Println("messages")
	//	fmt.Println(messages)
}

//Get return the value for given range (from, to)
//returns value starting with offset `from` to `to` (exclusive)
// Must: from < to
func Get(from, to uint64) []Message {
	// 2, 5 => 2,3,4
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
		fmt.Printf(" %d: %s ", i, msg)

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

	go func(messageConfig config.Message) {
		for {
			time.Sleep(messageConfig.MessageFlushIntervalInMilliSeconds * time.Millisecond)

			//keep looping on the above sleep interval when the message size is zero
			if len(messages) == 0 {
				continue
			}

			count, err := WriteToLogFile(messageConfig, messages)
			if err != nil {
				log.Error().Err(err).Msg("failed to write the messages to file")
				continue
			}
			log.Debug().Int("Count", count).Msg("Wrote %d messages to file")

		}
	}(messageConfig)
}
