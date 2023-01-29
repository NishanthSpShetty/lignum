package cluster

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var errUnexpectedResponse = errors.New("Unexpected response code: 200")

func Test_CreateSession(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	consulConfig := config.Consul{
		Host: "localhost",
		Port: 8500,
	}

	// test create session
	type args struct {
		config                    config.Consul
		sessionRenewalChannelChan chan struct{}
	}

	testCases := []struct {
		name       string
		args       args
		sessionId  string
		err        error
		controller *ConsulClusterController
	}{
		{
			name: "Create session successfully",
			args: args{
				sessionRenewalChannelChan: make(chan struct{}),
				config:                    consulConfig,
			},
			err:       nil,
			sessionId: "session-id",
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("CreateSession").Return("session-id", nil)
				return newMockClient(mclient)
			}(),
		},
		{
			name: "Failed to create session",
			args: args{
				sessionRenewalChannelChan: make(chan struct{}),
				config:                    consulConfig,
			},
			err:       errUnexpectedResponse,
			sessionId: "",
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("CreateSession").Return("", errUnexpectedResponse)
				return newMockClient(mclient)
			}(),
		},
	}

	for _, tt := range testCases {

		activeGoRoutineBefore := runtime.NumGoroutine()

		renewalChannel := make(chan struct{})
		err := tt.controller.CreateSession(consulConfig, renewalChannel)

		assert.ErrorIsf(t, err, tt.err, "CreateSession: %s", tt.name)

		if tt.sessionId != "" && tt.sessionId != tt.controller.SessionId {
			t.Fatalf("CreateSession: %s, Expected :%s, Got :%s", tt.name, tt.sessionId, tt.controller.SessionId)
		}
		close(renewalChannel)
		// give it a second to send signal to the routines
		time.Sleep(time.Millisecond)

		activeGoRoutineAfter := runtime.NumGoroutine()

		assert.Equalf(t, activeGoRoutineBefore, activeGoRoutineAfter, "CreateSession: GoRoutineLeakDetected: Active goroutine Before:%d, After:%d\n", activeGoRoutineBefore, activeGoRoutineAfter)
	}
}

func Test_AcquireLock(t *testing.T) {
	node := types.Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	type args struct {
		node       types.Node
		serviceKey string
	}

	testCases := []struct {
		name       string
		args       args
		acquired   bool
		err        error
		controller types.ClusterController
	}{
		{
			name: "lock acquired successfully",
			args: args{
				node:       node,
				serviceKey: "service/lignum/key/master",
			},
			acquired: true,
			err:      nil,

			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("AcquireLock").Return(true, nil)
				controller := newMockClient(mclient)
				controller.SessionId = "test-session-id"
				return controller
			}(),
		},
		{
			name: "failed to acquire lock on the consul",
			args: args{
				node:       node,
				serviceKey: "service/lignum/key/master",
			},
			acquired: false,
			err:      errUnexpectedResponse,

			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("AcquireLock").Return(false, errUnexpectedResponse)
				controller := newMockClient(mclient)
				controller.SessionId = "test-session-id"
				return controller
			}(),
		},
	}

	for _, tt := range testCases {
		acquired, err := tt.controller.AcquireLock(tt.args.node, tt.args.serviceKey)
		assert.ErrorIsf(t, err, tt.err, "AcquireLock: %s", tt.name)
		assert.Equalf(t, tt.acquired, acquired, "AcquireLock: %s", tt.name)
	}
}

func Test_GetLeader(t *testing.T) {
	node := types.Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	testservicekey := "test-service-key"
	testsessionid := "testsessionid"
	lockData := node.Json()

	type args struct {
		serviceKey string
	}

	testCases := []struct {
		name       string
		args       args
		want       types.Node
		err        error
		controller types.ClusterController
	}{
		{
			name: "consul api returns error",
			args: args{
				serviceKey: testservicekey,
			},
			want: types.Node{},
			err:  errUnexpectedResponse,
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("GetKVPair").Return(nil, errUnexpectedResponse)
				controller := newMockClient(mclient)
				controller.SessionId = testsessionid
				return controller
			}(),
		},
		{
			name: "consul returns nil kvpair",
			args: args{
				serviceKey: testservicekey,
			},
			want: types.Node{},
			err:  ErrLeaderNotFound,
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("GetKVPair").Return(nil, nil)
				controller := newMockClient(mclient)
				controller.SessionId = testsessionid
				return controller
			}(),
		},
		{
			name: "kvPair returned from consul is empty",
			args: args{},
			want: types.Node{},
			err:  ErrLeaderNotFound,
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}

				kvPair := &api.KVPair{
					Key:     testservicekey,
					Value:   lockData,
					Session: "",
				}
				mclient.On("GetKVPair").Return(kvPair, nil)
				controller := newMockClient(mclient)
				controller.SessionId = testsessionid
				return controller
			}(),
		},
		{
			name: "consul returns leader node",
			args: args{},
			want: node,
			err:  nil,
			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}

				kvPair := &api.KVPair{
					Key:     testservicekey,
					Value:   lockData,
					Session: testsessionid,
				}
				mclient.On("GetKVPair").Return(kvPair, nil)
				controller := newMockClient(mclient)
				controller.SessionId = "test-session-id"
				return controller
			}(),
		},
	}

	for _, tt := range testCases {
		got, err := tt.controller.GetLeader(tt.args.serviceKey)
		assert.ErrorIsf(t, err, tt.err, "GetLeader: %s", tt.name)
		assert.Equal(t, tt.want, got, "GetLeader: %s", tt.name)
	}
}
