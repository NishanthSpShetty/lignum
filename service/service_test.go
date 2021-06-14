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
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func createTestConfig() config.Config {
	return config.Config{
		Server:  config.Server{},
		Consul:  config.Consul{},
		Message: config.Message{},
	}

}

func Test_serviceStopAllGoroutine(t *testing.T) {
	//	preRoutine := runtime.NumGoroutine()
	config := createTestConfig()

	clusterController := &cluster.MockclusterController{
		ConsulClusterController: &cluster.ConsulClusterController{}}
	clusterController.On("CreateSession").Return(mock.Anything)
	clusterController.On("AcquireLock").Return(mock.Anything)
	clusterController.On("DestroySession").Return(mock.Anything)

	service := &Service{
		signalChannel:         make(chan os.Signal),
		ServiceId:             uuid.New().String(),
		Config:                config,
		ClusterController:     clusterController,
		ReplicationQueue:      make(chan message.Message, REPLICATION_QUEUE_SIZE),
		SessionRenewalChannel: make(chan struct{}),
		message:               message.New(config.Message),
		follower:              follower.New(),
	}
	service.apiServer = api.NewServer(service.ServiceId, service.ReplicationQueue, service.Config.Server, service.message, service.follower)

	go service.Start()
	close(service.signalChannel)
	//give it a second to kill all
	time.Sleep(time.Second)
	//(t, preRoutine, runtime.NumGoroutine(), "Number of go routine when service stopped will remain same beore startup")
}
