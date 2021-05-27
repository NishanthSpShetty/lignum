package service

import (
	"runtime"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	preRoutine := runtime.NumGoroutine()
	config := createTestConfig()

	clusterController := &cluster.MockclusterController{
		ConsulClusterController: &cluster.ConsulClusterController{}}
	clusterController.On("CreateSession").Return(mock.Anything)
	clusterController.On("AcquireLock").Return(mock.Anything)
	clusterController.On("DestroySession").Return(mock.Anything)

	service := &Service{
		ServiceId:             uuid.New().String(),
		Config:                config,
		ClusterController:     clusterController,
		ReplicationQueue:      make(chan message.MessageT, REPLICATION_QUEUE_SIZE),
		SessionRenewalChannel: make(chan struct{}),
	}
	go service.Start()

	service.Stop()
	time.Sleep(time.Millisecond)
	assert.Equal(t, preRoutine, runtime.NumGoroutine(), "Number of go routine when service stopped will remain same beore startup")
}
