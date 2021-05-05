package cluster

import (
	"errors"
	"runtime"
	"testing"
	"time"

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

	//test create session
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
		//give it a second to send signal to the routines
		time.Sleep(time.Millisecond)

		activeGoRoutineAfter := runtime.NumGoroutine()

		assert.Equalf(t, activeGoRoutineBefore, activeGoRoutineAfter, "CreateSession: GoRoutineLeakDetected: Active goroutine Before:%d, After:%d\n", activeGoRoutineBefore, activeGoRoutineAfter)

	}

	//	leaderNode, err := clusteController.GetLeader(serviceKey)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	if leaderNode.Port != 8080 {
	//		t.Fatal(err)
	//	}
	//
	//	if err := clusteController.DestroySession(); err != nil {
	//		t.Fatal(err)
	//	}

}

func Test_AquireLock(t *testing.T) {

	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	type args struct {
		node       Node
		serviceKey string
	}

	testCases := []struct {
		name       string
		args       args
		aquired    bool
		err        error
		controller ClusterController
	}{
		{
			name: "lock aquired successfully",
			args: args{
				node:       node,
				serviceKey: "service/lignum/key/master",
			},
			aquired: true,
			err:     nil,

			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("AquireLock").Return(true, nil)
				controller := newMockClient(mclient)
				controller.SessionId = "test-session-id"
				return controller
			}(),
		},
		{
			name: "failed to aquire lock on the consul",
			args: args{
				node:       node,
				serviceKey: "service/lignum/key/master",
			},
			aquired: false,
			err:     errUnexpectedResponse,

			controller: func() *ConsulClusterController {
				mclient := &mockConsulClient{}
				mclient.On("AquireLock").Return(false, errUnexpectedResponse)
				controller := newMockClient(mclient)
				controller.SessionId = "test-session-id"
				return controller
			}(),
		},
	}

	for _, tt := range testCases {
		acquired, err := tt.controller.AquireLock(tt.args.node, tt.args.serviceKey)
		assert.ErrorIsf(t, err, tt.err, "AquireLock: %s", tt.name)
		assert.Equalf(t, tt.aquired, acquired, "AquireLock: %s", tt.name)
	}
}

func Test_GetLeader(t *testing.T) {
	node := Node{
		Id:   "test-node",
		Host: "localhost",
		Port: 8080,
	}

	testservicekey := "test-service-key"
	testsessionid := "testsessionid"
	lockData, _ := node.Json()

	type args struct {
		serviceKey string
	}

	testCases := []struct {
		name       string
		args       args
		want       Node
		err        error
		controller ClusterController
	}{

		{
			name: "consul api returns error",
			args: args{
				serviceKey: testservicekey,
			},
			want: Node{},
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
			want: Node{},
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
			want: Node{},
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
