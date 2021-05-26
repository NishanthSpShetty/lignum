package cluster

import (
	"context"
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
	clusterController.On("AcquireLock", mock.Anything).Return()
	leaderElection(context.Background(), node, clusterController, serviceKey)
	assert.True(t, isLeader, "This node should be the leader")
}
