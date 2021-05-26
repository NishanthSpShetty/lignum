package cluster

import "github.com/stretchr/testify/mock"

type mockclusteController struct {
	*ConsulClusterController
	mock.Mock
}

func (m *mockclusteController) AcquireLock(node Node, serviceKey string) (bool, error) {
	return true, nil
}
