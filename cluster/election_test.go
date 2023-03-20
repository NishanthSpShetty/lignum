package cluster

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/wal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init_state() {
	state = &State{}
}

func Test_leaderStateUpdate(t *testing.T) {
	init_state()
	serviceKey := "service/lignum/key/master"
	node := cluster_types.Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	// unable to acquire lock
	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	clusterController.On("AcquireLock", mock.Anything).Return(false)
	leaderSignal := make(chan bool, 1)
	tryAcquireLock(node, clusterController, serviceKey, leaderSignal)
	assert.True(t, !state.isLeader(), "This node should is not the leader")

	// when lock acquired, we need to capture leader signal channel value, so spawn new routine.
	clusterController = &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	clusterController.On("AcquireLock", mock.Anything).Return(true)
	tryAcquireLock(node, clusterController, serviceKey, leaderSignal)
	assert.True(t, state.isLeader(), "This node should be the leader")
	// NOTE: we are not handling when channel write is not happening, as it would block indefinitely, on straight out read without any timeout on read implemented.
	assert.True(t, <-leaderSignal, "leader should send signal on the channel")

	close(leaderSignal)
}

func Test_LeaderElection(t *testing.T) {
	init_state()
	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	serviceKey := "service/lignum/key/master"
	node := cluster_types.Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}
	clusterController.On("AcquireLock", mock.Anything).Return(true)
	leaderSignal := make(chan bool, 2)
	go leaderElection(context.Background(), node, clusterController, serviceKey, 1*time.Second, leaderSignal)
	<-leaderSignal
	close(leaderSignal)
	assert.True(t, state.isLeader(), "This node should be the leader")
}

func Test_ConnectToLeader(t *testing.T) {
	init_state()
	fr := &types.FollowerRegistration{}
	mockServer := httptest.NewServer( /* handle: "api/follower/register" */

		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			r, _ := io.ReadAll(req.Body)
			er := json.Unmarshal(r, fr)
			assert.Nil(t, er, "unmarshal request to follower registration")
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}))

	clusterController := &MockclusterController{ConsulClusterController: &ConsulClusterController{}}
	serviceKey := "service/lignum/key/master"

	node := cluster_types.Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}
	mockURL, _ := url.Parse(mockServer.URL)
	port, _ := strconv.Atoi(mockURL.Port())
	leaderNode := cluster_types.Node{
		Id:   "leader-node",
		Host: mockURL.Hostname(),
		Port: port,
	}
	clusterController.On("GetLeader", mock.Anything).Return(leaderNode)

	walChannel := make(chan wal.Payload, 10)
	msg := message.New(config.Message{InitialSizePerTopic: 10}, walChannel)

	msg.Put(context.Background(), "leader_connect", []byte(" this is a leader connect message"))
	connectToLeader(context.Background(), serviceKey, clusterController, node, *http.DefaultClient, msg)

	assert.True(t, state.isConnectedLeader(), "should connect to leader")
	assert.Equal(t, leaderNode, *state.getLeader(), "should set the leader in cluster state")
	assert.Equal(t, node, fr.Node, "should recieve the current node as follower")
	assert.Equal(t, 1, len(fr.MessageStat), "should recieve message stat for  only 1 topic")
	assert.Equal(t, types.MessageStat{
		Topic:  "leader_connect",
		Offset: 1,
	}, fr.MessageStat[0], "should recieve message stat topic")
}
