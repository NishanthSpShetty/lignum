package service

import (
	"context"
	"os"

	_ "net/http/pprof"

	"github.com/NishanthSpShetty/lignum/api"
	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

//QUEUE_SIZE replication message queue size
const REPLICATION_QUEUE_SIZE = 1024

type Service struct {
	signalChannel         chan os.Signal
	Config                config.Config
	ServiceId             string
	SessionRenewalChannel chan struct{}
	ClusterController     cluster.ClusterController
	ReplicationQueue      chan message.MessageT
	Cancels               []context.CancelFunc
	apiServer             *api.Server
}

func New(config config.Config) (*Service, error) {

	consulClusterController, err := cluster.InitialiseClusterController(config.Consul)

	if err != nil {
		return nil, errors.Wrap(err, "Service.New")
	}

	s := &Service{
		signalChannel:         make(chan os.Signal),
		ServiceId:             uuid.New().String(),
		Config:                config,
		ClusterController:     consulClusterController,
		ReplicationQueue:      make(chan message.MessageT, REPLICATION_QUEUE_SIZE),
		SessionRenewalChannel: make(chan struct{}),
	}
	s.apiServer = api.NewServer(s.ServiceId, s.ReplicationQueue, s.Config.Server)
	return s, nil
}

//startClusterService Start all cluster management related routines
func (s *Service) startClusterService(ctx context.Context) error {

	err := s.ClusterController.CreateSession(s.Config.Consul, s.SessionRenewalChannel)
	if err != nil {
		return errors.Wrap(err, "Service.startClusterService")
	}
	//Start leader election routine
	cluster.InitiateLeaderElection(ctx, s.Config.Server, s.ServiceId, s.ClusterController)

	//connect to leader
	cluster.ConnectToLeader(s.Config.Server, s.ServiceId, s.ClusterController)
	return nil
}

func (s *Service) Start() error {

	log.Info().Str("ServiceID", s.ServiceId).Msg("Starting lignum - distributed messaging service")

	err := s.startClusterService(context.Background())
	if err != nil {
		return err
	}
	s.signalHandler()

	//initialize the message data structure
	message.Init(s.Config.Message)

	//start service routines
	//	message.StartFlusher(s.Config.Message)
	//	message.StartReplicator(s.ReplicationQueue)

	//once the cluster is setup we should be able start api service

	return s.apiServer.Serve()
}
