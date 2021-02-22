package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/lignum/config"
	log "github.com/sirupsen/logrus"
)

type ConsulClusterController struct {
	client    *api.Client
	SessionId string
}

func InitialiseClusterController(consulConfig config.Consul) (ClusterController, error) {

	var err error

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", consulConfig.Host, consulConfig.Port)

	client, err := api.NewClient(config)

	if err != nil {
		return &ConsulClusterController{}, err
	}
	return &ConsulClusterController{client: client}, nil
}

func (c *ConsulClusterController) renewSessionPeriodically(sessionId string, ttlS string, sessionRenewalChannel chan struct{}) {
	defer close(sessionRenewalChannel)
	for {
		c.client.Session().RenewPeriodic(ttlS, sessionId, nil, sessionRenewalChannel)
		time.Sleep(90 * time.Second)
	}
}

func (c *ConsulClusterController) CreateSession(consulConfig config.Consul, sessionRenewalChannel chan struct{}) error {

	sessionEntry := &api.SessionEntry{
		Name:      consulConfig.ServiceName,
		TTL:       consulConfig.SessionTTL,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, queryDuration, err := c.client.Session().Create(sessionEntry, nil)

	if err != nil {
		return err
	}
	log.Debugf("Consul session created, ID : %v, Aquired in :%dms  ", sessionId, queryDuration.RequestTime.Milliseconds())
	//TODO : should it be started conditionally?
	go c.renewSessionPeriodically(sessionId, consulConfig.SessionRenewalTTL, sessionRenewalChannel)
	c.SessionId = sessionId
	return err
}

func (c *ConsulClusterController) DestroySession() error {
	_, err := c.client.Session().Destroy(c.SessionId, nil)
	return err
}

func (c ConsulClusterController) GetLeader(serviceKey string) (NodeConfig, error) {
	kv, _, err := c.client.KV().Get(serviceKey, nil)
	nodeConfig := NodeConfig{}
	if err != nil {
		return nodeConfig, err
	}

	if kv == nil || kv.Session == "" {
		return nodeConfig, errLeaderNotFound
	}
	err = json.Unmarshal(kv.Value, &nodeConfig)
	return nodeConfig, err
}

func (c ConsulClusterController) AquireLock(nodeConfig NodeConfig, serviceKey string) (bool, time.Duration, error) {

	lockData, err := nodeConfig.json()

	if err != nil {
		return false, 0, err
	}

	kvPair := &api.KVPair{
		Key:     serviceKey,
		Value:   lockData,
		Session: c.SessionId,
	}
	acquired, writeMeta, err := c.client.KV().Acquire(kvPair, nil)
	if err != nil {
		return false, 0, err
	}
	return acquired, writeMeta.RequestTime, err

}
