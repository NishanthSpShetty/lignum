package cluster

import (
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

type mockConsulClient struct {
	mock.Mock
}

func (c *mockConsulClient) GetKVPair(serviceKey string) (*api.KVPair, error) {
	args := c.Called()
	obj := args.Get(0)
	if obj == nil {
		return nil, args.Error(1)

	}
	return obj.(*api.KVPair), args.Error(1)
}

func (c *mockConsulClient) AquireLock(kvPair *api.KVPair) (bool, *api.WriteMeta, error) {
	args := c.Called()
	return args.Bool(0), &api.WriteMeta{
		RequestTime: 100,
	}, args.Error(1)
}

func (c *mockConsulClient) RenewPeriodic(initialTTL string, id string, q *api.WriteOptions, doneCh <-chan struct{}) error {
	return nil
}

func (c *mockConsulClient) CreateSession(se *api.SessionEntry, q *api.WriteOptions) (string, *api.WriteMeta, error) {
	args := c.Called()
	return args.String(0), &api.WriteMeta{
		RequestTime: 100,
	}, args.Error(1)
}

func (c *mockConsulClient) DestroySession(sessionId string) error {
	return nil
}

func newMockClient(mclient *mockConsulClient) *ConsulClusterController {
	return &ConsulClusterController{
		client: mclient,
	}
}
