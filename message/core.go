package message

import (
	log "github.com/sirupsen/logrus"
)

type MessageT map[string]string

var message MessageT

func Init(initialLogSize int64) {
	//initialize the message
	log.Infof("Initializing message map with the initial log size %d", initialLogSize)
	message = make(MessageT, initialLogSize)
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
