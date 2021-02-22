package cluster

import (
	"testing"

	"github.com/lignum/config"
)

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
	nodeConfig := NodeConfig{
		NodeId: "test-node",
		NodeIp: "localhost",
		Port:   8080,
	}
	_, _, err = clusteController.AquireLock(nodeConfig, serviceKey)

	if err != nil {
		t.Fatal(err)
	}

	leaderNodeConfig, err := clusteController.GetLeader(serviceKey)
	if err != nil {
		t.Fatal(err)
	}

	if leaderNodeConfig.Port != 8080 {
		t.Fatal(err)
	}

	if err := clusteController.DestroySession(); err != nil {
		t.Fatal(err)
	}

}
