package cluster

import (
	"fmt"
	"strconv"
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

func (c *ConsulClusterController) renewSessionPeriodicall(sessionId string, ttlS string, sessionRenewalChannel chan struct{}) {
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
	go c.renewSessionPeriodicall(sessionId, consulConfig.SessionRenewalTTL, sessionRenewalChannel)
	c.SessionId = sessionId
	return err
}

func (c *ConsulClusterController) DestroySession() error {
	_, err := c.client.Session().Destroy(c.SessionId, nil)
	return err
}

func (c ConsulClusterController) GetLeader(serviceKey string) (*Leader, error) {
	kv, _, err := c.client.KV().Get(serviceKey, nil)
	if err != nil {
		return nil, err
	}

	if kv == nil || kv.Session == "" {
		//read the data we need and create Leader data
		return nil, errLeaderNotFound
	}
	port, err := strconv.Atoi(string(kv.Value))
	if err != nil {
		return nil, err
	}
	return &Leader{Port: port}, err

}

func (c ConsulClusterController) AquireLock(nodeConfig NodeConfig, serviceKey string) (bool, time.Duration, error) {
	kvPair := &api.KVPair{
		Key:     serviceKey,
		Value:   []byte(fmt.Sprintf("%d", nodeConfig.Port)),
		Session: c.SessionId,
	}
	acquired, writeMeta, err := c.client.KV().Acquire(kvPair, nil)
	return acquired, writeMeta.RequestTime, err

}
