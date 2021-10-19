package service

import (
	"os"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/api"
	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func createTestConfig() config.Config {
	return config.Config{
		Server: config.Server{},
		Consul: config.Consul{
			LeaderElectionIntervalInMilliSeconds: 100,
		},
		Message: config.Message{
			DataDir: "",
		},
		Follower: config.Follower{
			RegistrationOrLeaderCheckIntervalInSeconds: 1,
			HealthCheckIntervalInSecond:                1,
		},
		Wal: config.Wal{QueueSize: 10},
	}

}

func Test_serviceStopAllGoroutine(t *testing.T) {
	//	preRoutine := runtime.NumGoroutine()
	config := createTestConfig()

	clusterController := &cluster.MockclusterController{
		ConsulClusterController: &cluster.ConsulClusterController{}}
	clusterController.On("CreateSession").Return(mock.Anything)
	clusterController.On("CreateSession").Return(mock.Anything)
	clusterController.On("AcquireLock").Return(true)
	clusterController.On("DestroySession").Return(mock.Anything)

	walChannel := make(chan wal.Payload, config.Wal.QueueSize)

	service := &Service{
		signalChannel:         make(chan os.Signal),
		ServiceId:             uuid.New().String(),
		Config:                config,
		ClusterController:     clusterController,
		ReplicationQueue:      make(chan replication.Payload, REPLICATION_QUEUE_SIZE),
		SessionRenewalChannel: make(chan struct{}),
		message:               message.New(config.Message, walChannel),
		followerRegistry:      follower.New(),
	}
	service.wal = wal.New(config.Wal, config.Message.DataDir, walChannel)
	service.replicator = replication.New(service.ReplicationQueue, service.followerRegistry)
	service.apiServer = api.NewServer(service.ServiceId, service.ReplicationQueue, service.Config.Server, service.message, service.followerRegistry)

	go service.Start()
	close(service.signalChannel)
	//give it a second to kill all
	time.Sleep(time.Second)
	//(t, preRoutine, runtime.NumGoroutine(), "Number of go routine when service stopped will remain same beore startup")
}
