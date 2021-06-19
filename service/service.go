package service

import (
	"context"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/NishanthSpShetty/lignum/api"
	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
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
	ReplicationQueue      chan message.Message
	Cancels               []context.CancelFunc
	apiServer             *api.Server
	message               *message.MessageStore
	follower              *follower.FollowerRegistry
	running               bool
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
		ReplicationQueue:      make(chan message.Message, REPLICATION_QUEUE_SIZE),
		SessionRenewalChannel: make(chan struct{}),
		message:               message.New(config.Message),
		follower:              follower.New(),
	}
	s.apiServer = api.NewServer(s.ServiceId, s.ReplicationQueue, s.Config.Server, s.message, s.follower)
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
	interval := 1 * time.Second
	cluster.FollowerRegistrationRoutine(ctx, s.Config.Server, interval, s.ServiceId, s.ClusterController)
	return nil
}

func (s *Service) SetStarted() {
	s.running = true
}

func (s *Service) SetStopped() {
	s.running = false
}

func (s *Service) Stopped() bool {
	return !s.running
}

func (s *Service) addCancel(fn context.CancelFunc) {
	s.Cancels = append(s.Cancels, fn)
}

func (s *Service) Start() error {

	log.Info().Str("ServiceID", s.ServiceId).Msg("starting lignum - distributed messaging service")

	ctx, cancel := context.WithCancel(context.Background())
	s.addCancel(cancel)
	err := s.startClusterService(ctx)
	if err != nil {
		return err
	}
	s.signalHandler()

	//start service routines
	s.follower.StartHealthCheck(ctx, 1*time.Second)
	//	message.StartFlusher(s.Config.Message)
	//	message.StartReplicator(s.ReplicationQueue)

	//mark service as running
	s.SetStarted()
	//once the cluster is setup we should be able start api service
	return s.apiServer.Serve()
}
