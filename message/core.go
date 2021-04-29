package message

import (
	"time"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

var count = 0

type MessageT struct {
	Id      int
	Message string
}

var messages []MessageT

func Init(messageConfig config.Message) {
	log.Infof("Initializing messages with config %v", messageConfig)

	//	message = make(MessageT)
	messages = ReadFromLogFile(messageConfig.MessageDir)
	count = len(messages)
}

func Put(msg string) {
	messages = append(messages, MessageT{count, msg})
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
				log.Errorf("failed to write the messages to file : %v", err.Error())
				continue
			}
			log.Debugf("Wrote %d messages to file", count)

		}
	}(messageConfig)
}
