package message

import "github.com/rs/zerolog/log"

//StartReplicator start replication routine to replicate the messages to all nodes
func StartReplicator(messageChannel chan Message) {

	log.Info().Msg("Replicator service is running..")
	go func(messageChannel <-chan Message) {
		for msg := range messageChannel {
			log.Debug().Interface("Message", msg).Msg("Recieved message for replication ")
		}
	}(messageChannel)
}
