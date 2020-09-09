package cluster

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

//isLeader mark it as true when the current node becomes the leader
var isLeader = false

func leaderElection(nodeConfig NodeConfig, sessionId string, serviceKey string) {
	loggedOnce := false
	kvPair := &api.KVPair{
		Key:     serviceKey,
		Value:   []byte(fmt.Sprintf("%d", nodeConfig.Port)),
		Session: sessionId,
	}
	//start polling to aquire the lock indefinitely
	for {
		if isLeader {
			//if the current node is leader, stop the busy loop for now
			return
		}
		aquired, queryDuration, err := client.KV().Acquire(kvPair, nil)
		if err != nil {
			log.Errorf("Failed to aquire lock")
			continue
		}
		if aquired {
			isLeader = aquired
			log.Infof("Lock aquired and marking the node  as leader, Took : %dms\n",
				queryDuration.RequestTime.Milliseconds())
		} else {

			if !loggedOnce {
				log.Debug("Lock is already taken, will check again...")
				loggedOnce = true
			}
			//every 10ms attempt to grab a lock on the consul.
			time.Sleep(10 * time.Millisecond)
		}

	}
}
