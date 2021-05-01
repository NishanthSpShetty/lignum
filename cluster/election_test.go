package cluster

import (
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
)

func Test_LeaderElection(t *testing.T) {
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
	leaderElection(node, clusteController, serviceKey)
	//sleep for 10ms,
	time.Sleep(10 * time.Millisecond)
	leader, err := clusteController.GetLeader(serviceKey)

	if err != nil {
		t.Fatal(err)
	}

	if leader.Port != 8080 {
		t.Fatal(err)
	}

	if err := clusteController.DestroySession(); err != nil {
		t.Fatal(err)
	}

}
