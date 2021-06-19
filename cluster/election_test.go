package cluster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init_state() {
	state = &State{}
}

func Test_leaerStateUpdate(t *testing.T) {
	init_state()
	serviceKey := "service/lignum/key/master"
	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	//unable to acquire lock
	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	clusterController.On("AcquireLock", mock.Anything).Return(false)
	tryAcquireLock(node, clusterController, serviceKey)
	assert.True(t, !state.isLeader(), "This node should is not the leader")

	//when lock acquired
	clusterController = &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	clusterController.On("AcquireLock", mock.Anything).Return(true)
	tryAcquireLock(node, clusterController, serviceKey)
	assert.True(t, state.isLeader(), "This node should be the leader")
}

func Test_LeaderElection(t *testing.T) {

	init_state()
	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	serviceKey := "service/lignum/key/master"
	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}
	clusterController.On("AcquireLock", mock.Anything).Return(true)
	leaderElection(context.Background(), node, clusterController, serviceKey)
	assert.True(t, state.isLeader(), "This node should be the leader")
}

func Test_ConnectToLeader(t *testing.T) {

	init_state()
	mockServer := httptest.NewServer( /* handle: "api/follower/register" */

		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}))

	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	serviceKey := "service/lignum/key/master"

	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}
	thisNodeData, _ := node.Json()

	mockURL, _ := url.Parse(mockServer.URL)
	port, _ := strconv.Atoi(mockURL.Port())
	leaderNode := Node{
		Id:   "leader-node",
		Host: mockURL.Hostname(),
		Port: port,
	}
	clusterController.On("GetLeader", mock.Anything).Return(leaderNode)
	connectToLeader(serviceKey, clusterController, thisNodeData, *http.DefaultClient)

	assert.True(t, state.isConnectedLeader(), "should connect to leader")
	assert.Equal(t, leaderNode, *state.getLeader(), "should set the leader in cluster state")
}
