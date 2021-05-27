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
	m.Called()
	return true, nil
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
