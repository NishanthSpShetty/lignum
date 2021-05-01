package cluster

import (
	"testing"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog"
)

func Test_ClusterController(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	consulConfig := config.Consul{
		Host: "localhost",
		Port: 8500,
	}

	controller := newMockClient(&mockConsulClient{})

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
				mclient.On("CreateSession").Return("session-id")
				return newMockClient(mclient)
			}(),
		},
	}

	for _, tt := range testCases {

		err := tt.controller.CreateSession(consulConfig, make(chan struct{}))
		if err != tt.err {
			t.Fatalf("CreateSession: %s, Expected %v, Got %v", tt.name, tt.err, err)
		}
		if tt.sessionId != "" && tt.sessionId != tt.controller.SessionId {
			t.Fatalf("CreateSession: %s, Expected %s, Got %s", tt.name, tt.sessionId, controller.SessionId)
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
