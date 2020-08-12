package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
)

const (
	Hello = "hello"
	//Register this key on consul and get  a lock on it.
	ServiceKey  = "service/disqlogger/key/master"
	ServiceName = "distributed-quote-logger"
	//Standard port where all follower should register with primary server.
	PrimaryPort = 9090
	ttlS        = "10s"
)

func main() {
	ch := make(chan int)
	log.SetLevel(log.DebugLevel)

	log.Info("Starting service..")
	config := api.DefaultConfig()

	config.Address = "localhost:8500"

	client, err := api.NewClient(config)

	if err != nil {
		log.Panic(err)
		return
	}

	//create session and become master.
	sessionEntry := &api.SessionEntry{
		Name:      ServiceName,
		TTL:       ttlS,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, queryDuration, err := client.Session().Create(sessionEntry, nil)

	if err != nil {
		log.Errorf("Failed to create session %v \n", err)
		return
	}

	log.Debugf("Consul session created, ID : %v, Aquired in :%d  ", sessionId, queryDuration.RequestTime.Milliseconds)

	isLeader := false
	//aquire lock and make it as a leader.
	go func() {
		for {
			if !isLeader {

				isLeader, queryDuration, err = client.KV().Acquire(&api.KVPair{
					Key:     ServiceKey,
					Value:   []byte(fmt.Sprintf("%d", PrimaryPort)),
					Session: sessionId,
				}, nil)

				//				isLeader = isLeaderShadow

				if err != nil {
					log.Errorf("Failed to aquire lock %v \n", err)
					return

				}

				if isLeader {
					log.Infof("Lock aquired and marking the service as leader, Lock aquired in %dms\n", queryDuration.RequestTime.Milliseconds)
				} else {
					log.Debug("Lock is already taken, will check again...")
				}

				time.Sleep(60 * time.Millisecond)

			}
		}
	}()

	//Create a session renewer routine.
	doneChan := make(chan struct{})
	defer close(doneChan)
	go func() {
		for {

			client.Session().RenewPeriodic(ttlS, sessionId, nil, doneChan)
			time.Sleep(90 * time.Second)
		}
	}()

	ch <- 1
}
