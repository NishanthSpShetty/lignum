package message

import log "github.com/sirupsen/logrus"

//StartReplicator start replication routine to replicate the messages to all nodes
func StartReplicator(messageChannel chan MessageT) {

	log.Infoln("Replicator service is running..")
	go func(messageChannel <-chan MessageT) {
		for {

			messag := <-messageChannel
			log.Debugln("Recieved message for replication ", messag)

		}
	}(messageChannel)
}
