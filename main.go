package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/lignum/api"
	"github.com/lignum/cluster"
	"github.com/lignum/config"
)

func signalHandler(sessionRenewalChannel chan struct{}, serviceId string, clusteController cluster.ClusterController) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	//read the signal and discard
	<-signalChannel
	log.Infof("Destroying session and stopping service [ServiceID : %s ]\n", serviceId)
	close(sessionRenewalChannel)

	err := clusteController.DestroySession()

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
	appConfig.SetServiceId(serviceId)
	log.Infof("Starting lignum - distributed messaging service service [ServiceID : %s ].\n", serviceId)

	sessionRenewalChannel := make(chan struct{})
	consulClusterController, err := cluster.InitialiseClusterController(appConfig.Consul)
	if err != nil {
		log.Error("Failed to initialise the consul client", err)
		return
	}

	log.Infof("Loaded app config %v ", appConfig)
	err = consulClusterController.CreateSession(appConfig.Consul, sessionRenewalChannel)
	if err != nil {
		log.Errorf("Failed to create the consul session %v, Check if the consul is running and reachable.", err)
		return
	}

	//Start leader election routine
	cluster.InitiateLeaderElection(appConfig.Server, serviceId, consulClusterController)
	go signalHandler(sessionRenewalChannel, serviceId, consulClusterController)

	//connect to leader
	cluster.ConnectToLeader(appConfig.Server, serviceId, consulClusterController)
	//once the cluster is setup we should be able start api service
	api.StartApiService(appConfig, serviceId)
}
