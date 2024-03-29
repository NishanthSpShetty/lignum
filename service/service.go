package service

import (
	"context"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/NishanthSpShetty/lignum/api"
	cluster "github.com/NishanthSpShetty/lignum/cluster"
	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	c "github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// QUEUE_SIZE replication message queue size
const (
	REPLICATION_QUEUE_SIZE = 1024
	FOLLOWER_QUEUE_SIZE    = 12
)

type Service struct {
	signalChannel         chan os.Signal
	Config                c.Config
	ServiceId             string
	SessionRenewalChannel chan struct{}
	ClusterController     cluster_types.ClusterController
	ReplicationQueue      chan replication.Payload
	Cancels               []context.CancelFunc
	apiServer             *api.Server
	message               *message.MessageStore
	followerRegistry      *follower.FollowerRegistry
	liveReplicator        *replication.LiveReplicator
	walReplicator         *replication.WALReplicator
	wal                   *wal.Wal
	walService            *WalService
	running               bool
	leaderSignal          chan bool
}

func New(config c.Config) (*Service, error) {
	consulClusterController, err := cluster.InitialiseClusterController(config.Consul)
	if err != nil {
		return nil, errors.Wrap(err, "Service.New")
	}
	walChannel := make(chan wal.Payload, config.Wal.QueueSize)
	leaderSignal := make(chan bool, 2)
	followerQueue := make(chan *follower.Follower, FOLLOWER_QUEUE_SIZE)

	s := &Service{
		signalChannel:         make(chan os.Signal),
		ServiceId:             uuid.New().String(),
		Config:                config,
		ClusterController:     consulClusterController,
		ReplicationQueue:      make(chan replication.Payload, config.Replication.InternalQueueSize),
		SessionRenewalChannel: make(chan struct{}),
		message:               message.New(config.Message, walChannel),
		followerRegistry:      follower.New(followerQueue),
		leaderSignal:          leaderSignal,
	}
	s.wal = wal.New(config.Wal, config.Message.DataDir, walChannel)
	s.walService = NewReplicaionService(config, s.message)
	s.liveReplicator = replication.NewLiveReplicator(s.ReplicationQueue, s.followerRegistry)

	s.walReplicator = replication.NewWALReplication(followerQueue, leaderSignal, config.Replication.WalSyncIntervalInSec)

	server, err := api.NewServer(s.ServiceId, s.ReplicationQueue, s.Config.Server, s.message, s.followerRegistry)
	if err != nil {
		return nil, errors.Wrap(err, "Service.New")
	}
	s.apiServer = server
	return s, nil
}

// startClusterService Start all cluster management related routines
func (s *Service) startClusterService(ctx context.Context) error {
	err := s.ClusterController.CreateSession(s.Config.Consul, s.SessionRenewalChannel)
	if err != nil {
		return errors.Wrap(err, "Service.startClusterService")
	}
	// Start leader election routine
	cluster.InitiateLeaderElection(ctx, s.Config, s.ServiceId, s.ClusterController, s.leaderSignal)

	// connect to leader
	cluster.FollowerRegistrationRoutine(ctx, s.Config, s.ServiceId, s.ClusterController, s.message)
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
	healthCheckInterval := s.Config.Follower.HealthCheckIntervalInSecond * time.Second
	healthCheckTimeout := s.Config.Follower.HealthCheckTimeoutInMilliSeconds * time.Millisecond
	clientTimeout := s.Config.Replication.ClientTimeoutInMilliSeconds * time.Millisecond

	log.Info().Str("service_id", s.ServiceId).Msg("starting lignum - distributed messaging service")

	ctx, cancel := context.WithCancel(context.Background())
	s.addCancel(cancel)
	s.walReplicator.Start(ctx, clientTimeout, s.message)
	err := s.startClusterService(ctx)
	if err != nil {
		return err
	}
	s.signalHandler()

	// start service routines
	s.followerRegistry.StartHealthCheck(ctx, healthCheckInterval, healthCheckTimeout)
	s.liveReplicator.Start(ctx, clientTimeout)
	s.wal.StartWalWriter(ctx)
	s.walService.Start(ctx)

	s.message.RestoreWAL(s.wal)
	// mark service as running
	s.SetStarted()
	// once the cluster is setup we should be able start api service
	return s.apiServer.Start()
}
