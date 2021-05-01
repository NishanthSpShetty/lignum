package cluster

import (
	"testing"

	"github.com/lignum/config"
	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func Test_ClusterController(t *testing.T) {
	consulConfig := config.Consul{
		Host: "localhost",
		Port: 8500,
	}

	clusteController, err := InitialiseClusterController(consulConfig)
	if err != nil {
		t.Fatal(err)
	}

	if err := clusteController.CreateSession(consulConfig, make(chan struct{})); err != nil {
		t.Fatalf("Failed to create session %v \n", err)
	}
	serviceKey := "service/lignum/key/master"
	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}
	_, _, err = clusteController.AquireLock(node, serviceKey)

	if err != nil {
		t.Fatal(err)
	}

	leaderNode, err := clusteController.GetLeader(serviceKey)
	if err != nil {
		t.Fatal(err)
	}

	if leaderNode.Port != 8080 {
		t.Fatal(err)
	}

	if err := clusteController.DestroySession(); err != nil {
		t.Fatal(err)
	}

}
