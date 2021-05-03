package cluster

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog"
)

var errUnexpectedResponse = errors.New("Unexpected response code: 200")

func Test_ClusterController(t *testing.T) {
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

		if !errors.Is(tt.err, err) {
			t.Fatalf("CreateSession: %s, Expected :%v, Got :%v", tt.name, tt.err, err)
		}

		if tt.sessionId != "" && tt.sessionId != tt.controller.SessionId {
			t.Fatalf("CreateSession: %s, Expected :%s, Got :%s", tt.name, tt.sessionId, tt.controller.SessionId)
		}
		close(renewalChannel)
		//give it a second to send signal to the routines
		time.Sleep(time.Millisecond)

		activeGoRoutineAfter := runtime.NumGoroutine()

		if activeGoRoutineAfter != activeGoRoutineBefore {
			t.Fatalf("CreateSession: GoRoutineLeakDetected: Active goroutine Before:%d, After:%d\n", activeGoRoutineBefore, activeGoRoutineAfter)

		}

	}

	//	serviceKey := "service/lignum/key/master"
	//	node := Node{
	//		Id:   "test-node",
	//		Host: "localhost",
	//		Port: 8080,
	//	}
	//	_, _, err = clusteController.AquireLock(node, serviceKey)
	//
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
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
