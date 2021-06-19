package cluster

import (
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/stretchr/testify/mock"
)

type MockclusterController struct {
	*ConsulClusterController
	mock.Mock
}

func (m *MockclusterController) AcquireLock(node Node, serviceKey string) (bool, error) {
	args := m.Called()

	return args.Bool(0), nil
}

func (m *MockclusterController) CreateSession(consulConfig config.Consul, sessionRenewalChannel chan struct{}) error {
	//	m.SessionId = "DummySessionId"

	m.Called()
	return nil
}

func (m *MockclusterController) DestroySession() error {
	m.Called()
	return nil
}

func (m *MockclusterController) GetLeader(serviceKey string) (Node, error) {
	args := m.Called()
	return args.Get(0).(Node), nil
}
