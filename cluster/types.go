package cluster

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

var (
	ErrLeaderNotFound = errors.New("failed to get leader from the cluster controller")
)

type IClient interface {
	GetKVPair(serviceKey string) (*api.KVPair, error)
	AquireLock(kvPair *api.KVPair) (bool, *api.WriteMeta, error)
	RenewPeriodic(initialTTL string, id string, q *api.WriteOptions, doneCh <-chan struct{}) error
	CreateSession(se *api.SessionEntry, q *api.WriteOptions) (string, *api.WriteMeta, error)
	DestroySession(sessionId string) error
}

type Client struct {
	*api.Client
}

func (c *Client) GetKVPair(serviceKey string) (*api.KVPair, error) {
	kvPair, _, err := c.Client.KV().Get(serviceKey, nil)
	return kvPair, err
}

func (c *Client) AquireLock(kvPair *api.KVPair) (bool, *api.WriteMeta, error) {
	return c.Client.KV().Acquire(kvPair, nil)
}

func (c *Client) RenewPeriodic(initialTTL string, id string, q *api.WriteOptions, doneCh <-chan struct{}) error {
	return c.Client.Session().RenewPeriodic(initialTTL, id, q, doneCh)
}

func (c *Client) CreateSession(se *api.SessionEntry, q *api.WriteOptions) (string, *api.WriteMeta, error) {
	return c.Client.Session().Create(se, q)
}

func (c *Client) DestroySession(sessionId string) error {
	_, err := c.Client.Session().Destroy(sessionId, nil)
	return err
}
