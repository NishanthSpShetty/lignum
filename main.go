package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/lignum/cluster"
	"github.com/lignum/config"
)

func signalHandler(sessionRenewalChannel chan struct{}, serviceId  string, sessionId string) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	//read the signal and discard
	<-signalChannel
	log.Infof("Destroying session and stopping service [ServiceID : %s ]\n", serviceId)
	close(sessionRenewalChannel)

	err := cluster.DestroySession(sessionId)

	if err != nil {
		log.Error("Failed to destroy the session ", err)
	}
	os.Exit(0)
}

func main() {
	log.SetLevel(log.DebugLevel)

	configFile := flag.String("config", "", "get configuration from file")
	flag.Parse()

	appConfig, err := config.GetConfig(*configFile)
	if err != nil {
		log.Error("Failed to read config", err)
		return
	}

	serviceId := uuid.New().String()
	appConfig.SetServiceId( serviceId)
	log.Infof("Starting lignum - distributed messaging service service [ServiceID : %s ].\n", serviceId)

	sessionRenewalChannel := make(chan struct{})
	err = cluster.InitialiseConsulClient(appConfig.Consul)
	if err != nil {
		log.Error("Failed to initialise the consul client", err)
		return
	}

	log.Infof("Loaded app config %v ", appConfig)
	sessionId, err := cluster.CreateConsulSession(appConfig.Consul, sessionRenewalChannel)
	if err != nil {
		log.Error("Failed to create the consul session %v, Check if the consul is running and reachable.", err)
		return
	}

	//Start leader election routine
	cluster.InitiateLeaderElection(appConfig.Server, serviceId, sessionId)
	go signalHandler(sessionRenewalChannel, serviceId, sessionId )
	//connect to leader

	//TODO: try to connect to the leader, if not found call the leader election routine to make this service as the leader,
	//so we should start the leader connection routine.
	cluster.ConnectToLeader(appConfig.Server, sessionId)

	//start the work.
	log.Infof("Starting HTTP service at %s:%d \n", appConfig.Server.Host, appConfig.Server.Port)
	address := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	http.HandleFunc("/service/api/follower/register", func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Infof("Request received for follower registration %v ", string(requestBody))

		fmt.Fprintf(w, "Follower registered service  %s\n", serviceId)
	})
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "PONG")
	})
	log.Panic(http.ListenAndServe(address, nil))
}
