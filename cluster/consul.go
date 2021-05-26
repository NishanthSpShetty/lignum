package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

type ConsulClusterController struct {
	client    IClient
	SessionId string
}

//assert that ConsulClusterController implemnets ClusterController
var _ ClusterController = &ConsulClusterController{}

func InitialiseClusterController(consulConfig config.Consul) (*ConsulClusterController, error) {
	var err error
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", consulConfig.Host, consulConfig.Port)

	client, err := api.NewClient(config)

	if err != nil {
		return &ConsulClusterController{}, err
	}
	return &ConsulClusterController{client: &Client{client}}, nil
}

func (c *ConsulClusterController) renewSessionPeriodically(sessionId string, ttlS string, sessionRenewalChannel chan struct{}) {
	for {
		select {
		//if this channel is closed, stop this loop
		case <-sessionRenewalChannel:
			return

		case <-time.After(90 * time.Second):
			c.client.RenewPeriodic(ttlS, sessionId, nil, sessionRenewalChannel)
		}
	}
}

func (c *ConsulClusterController) CreateSession(consulConfig config.Consul, sessionRenewalChannel chan struct{}) error {

	sessionEntry := &api.SessionEntry{
		Name:      consulConfig.ServiceName,
		TTL:       consulConfig.SessionTTL,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, writeMeta, err := c.client.CreateSession(sessionEntry, nil)

	if err != nil {
		return err
	}
	log.Debug().
		Str("SessionId", sessionId).
		Str("Duration", writeMeta.RequestTime.String()).
		Msg("Consul session created")

	//TODO : should it be started conditionally?
	go c.renewSessionPeriodically(sessionId, consulConfig.SessionRenewalTTL, sessionRenewalChannel)
	c.SessionId = sessionId
	return err
}

func (c *ConsulClusterController) DestroySession() error {
	return c.client.DestroySession(c.SessionId)
}

func (c ConsulClusterController) GetLeader(serviceKey string) (Node, error) {
	kvPair, err := c.client.GetKVPair(serviceKey)
	nodeConfig := Node{}
	if err != nil {
		return nodeConfig, err
	}

	if kvPair == nil || kvPair.Session == "" {
		return nodeConfig, ErrLeaderNotFound
	}
	err = json.Unmarshal(kvPair.Value, &nodeConfig)
	return nodeConfig, err
}

func (c *ConsulClusterController) AcquireLock(node Node, serviceKey string) (bool, error) {

	lockData, err := node.Json()

	if err != nil {
		return false, err
	}

	kvPair := &api.KVPair{
		Key:     serviceKey,
		Value:   lockData,
		Session: c.SessionId,
	}

	acquired, writeMeta, err := c.client.AcquireLock(kvPair)
	if err != nil {
		return false, err
	}

	if acquired {
		log.Debug().
			Str("Duration", writeMeta.RequestTime.String()).
			RawJSON("Node", lockData).
			Bool("Acquired", acquired).
			Msg("Consul lock aquired on the session")
	}
	return acquired, err

}
