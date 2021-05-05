package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_LeaderElection(t *testing.T) {

	clusterController := &mockclusteController{ConsulClusterController: &ConsulClusterController{}}
	serviceKey := "service/lignum/key/master"
	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	clusterController.On("AquireLock", mock.Anything).Return()
	leaderElection(node, clusterController, serviceKey)

	assert.True(t, isLeader, "this node should be the leader")
}
