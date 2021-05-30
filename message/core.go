package message

import (
	"context"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

var count = 0

type Message struct {
	Id      int
	Message string
}

var messages []Message

func Init(messageConfig config.Message) {
	log.Info().
		Interface("MessageConfig", messageConfig).
		Msg("Initializing messages")

	//	message = make(MessageT)
	messages = ReadFromLogFile(messageConfig.MessageDir)
	count = len(messages)
}

func Put(ctx context.Context, msg string) {
	messages = append(messages, Message{count, msg})
}

func Get(from, to int) []string {
	return []string{}
}

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
