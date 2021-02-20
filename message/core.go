package message

import (
	"time"

	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

type MessageT map[string]string

var message MessageT
var messageChannel chan MessageT

func Init(messageConfig config.Message) {
	log.Infof("Initializing messages with config %v", messageConfig)

	message = make(MessageT)
	message = ReadFromLogFile(messageConfig.MessageDir)
}

func Put(key, value string) {
	message[key] = value
}

func Get(key string) string {
	v, ok := message[key]
	if ok {
		return v
	} else {
		return ""
	}
}

//StartFlusher start flusher routine to write the messages to file
func StartFlusher(messageConfig config.Message) {

	go func(messageConfig config.Message) {
		for {
			time.Sleep(messageConfig.MessageFlushIntervalInMilliSeconds * time.Millisecond)

			//keep looping on the above sleep interval when the message size is zero
			if len(message) == 0 {
				continue
			}

			err := WriteToLogFile(messageConfig, message)
			if err != nil {
				log.Errorf("failed to write the messages to file : %v", err.Error())
			}
		}
	}(messageConfig)
}
