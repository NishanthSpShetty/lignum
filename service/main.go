package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
)

const (
	//Register this key on consul and get  a lock on it.
	ServiceKey  = "service/disqlogger/key/master"
	ServiceName = "distributed-quote-logger"
	//Standard port where all follower should register with primary server.
	PrimaryPort = 9090
	ttlS        = "10s"
)

var isLeader = false
var serviceId uuid.UUID
var client *api.Client

func signalHandler(sessionRenewalChannel chan struct{}, sessionId string) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	//read the signal and discard
	<-signalChannel
	log.Infof("Destroying session and stopping service [ServiceID : %s ]", serviceId.String())
	close(sessionRenewalChannel)

	_, err := client.Session().Destroy(sessionId, nil)

	if err != nil {
		log.Error("Failed to destroy the session ", err)
	}
	os.Exit(0)
}

//leaderElection Function will keep trying to aquire lock on the `ServiceKey`
func leaderElection(sessionId string) {
	//make sure there is a leader registered on the consul all time.
	for {
		if !isLeader {

			aquired, queryDuration, err := client.KV().Acquire(&api.KVPair{
				Key:     ServiceKey,
				Value:   []byte(fmt.Sprintf("%d", PrimaryPort)),
				Session: sessionId,
			}, nil)

			if err != nil {
				log.Errorf("Failed to aquire lock %v \n", err)
				return

			}

			if aquired {
				isLeader = aquired

				log.Infof("Lock aquired and marking the service as leader, Lock aquired in %dms\n", queryDuration.RequestTime.Milliseconds)
			} else {
				log.Debug("Lock is already taken, will check again...")
			}

			time.Sleep(60 * time.Millisecond)

		} else {
			//if once the leader elected, stop the busy loop, will change this later if need to reaquire the lock\
			return
		}
	}
}

func main() {
	ch := make(chan int)

	serviceId = uuid.New()

	log.SetLevel(log.DebugLevel)

	log.Info("Starting loger service [ServiceID : %s ].", serviceId.String())
	config := api.DefaultConfig()

	config.Address = "localhost:8500"
	var err error
	client, err = api.NewClient(config)

	if err != nil {
		log.Error("Failed to create the new consul client ", err)
		return
	}

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

	//Start leader election routine
	go leaderElection(sessionId)

	doneChan := make(chan struct{})
	signalHandler(doneChan, sessionId)
	//Create a session renewer routine.
	go func() {
		defer close(doneChan)
		for {
			client.Session().RenewPeriodic(ttlS, sessionId, nil, doneChan)
			time.Sleep(90 * time.Second)
		}
	}()

	//start the work.
	ch <- 1
}
