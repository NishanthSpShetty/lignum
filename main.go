package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/NishanthSpShetty/lignum/api"
	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func signalHandler(sessionRenewalChannel chan struct{}, serviceId string, clusteController cluster.ClusterController) {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	//read the signal and discard
	<-signalChannel
	log.Info().
		Str("ServiceID", serviceId).
		Msg("Destroying session and stopping service ")
	close(sessionRenewalChannel)

	err := clusteController.DestroySession()

	if err != nil {
		log.Error().Err(err).Msg("Failed to destroy the session ")
	}
	os.Exit(0)
}

func initialiseLogger(development bool) {
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	if development {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	initialiseLogger(true)

	configFile := flag.String("config", "", "get configuration from file")
	flag.Parse()

	appConfig, err := config.GetConfig(*configFile)
	if err != nil {
		//o	 ("Failed to read config", err)
		log.Error().Err(err).Msg("Failed to read config")
		return
	}

	serviceId := uuid.New().String()
	appConfig.SetServiceId(serviceId)
	log.Info().Str("ServiceID", serviceId).Msg("Starting lignum - distributed messaging service")

	sessionRenewalChannel := make(chan struct{})
	consulClusterController, err := cluster.InitialiseClusterController(appConfig.Consul)

	if err != nil {
		log.Error().Err(err).Msg("Failed to initialise the consul client")
		return
	}

	log.Debug().Interface("Config", appConfig).Msg("Loaded app config")
	err = consulClusterController.CreateSession(appConfig.Consul, sessionRenewalChannel)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create the consul session.")
		return
	}

	//Start leader election routine
	cluster.InitiateLeaderElection(context.Background(), appConfig.Server, serviceId, consulClusterController)
	go signalHandler(sessionRenewalChannel, serviceId, consulClusterController)

	//connect to leader
	cluster.ConnectToLeader(appConfig.Server, serviceId, consulClusterController)
	//initialize the message data structure
	message.Init(appConfig.Message)
	messageChannel := make(chan message.MessageT)

	//start service routines
	message.StartFlusher(appConfig.Message)
	message.StartReplicator(messageChannel)

	//once the cluster is setup we should be able start api service
	api.StartApiService(appConfig, serviceId, messageChannel)
}
