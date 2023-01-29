package cluster

import (
	"encoding/json"
	"fmt"

	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

type ConsulClusterController struct {
	client    IClient
	SessionId string
}

// assert that ConsulClusterController implemnets ClusterController
var _ cluster_types.ClusterController = &ConsulClusterController{}

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

func (c *ConsulClusterController) renewSessionPeriodically(sessionId string, ttl int, sessionRenewalChannel chan struct{}) {
	// spawn go routine for renewal and return
	ttlS := fmt.Sprintf("%ds", ttl)
	go c.client.RenewPeriodic(ttlS, sessionId, nil, sessionRenewalChannel)
}

func (c *ConsulClusterController) CreateSession(consulConfig config.Consul, sessionRenewalChannel chan struct{}) error {
	sessionEntry := &api.SessionEntry{
		Name:      consulConfig.ServiceName,
		TTL:       fmt.Sprintf("%ds", consulConfig.SessionTTLInSeconds),
		LockDelay: consulConfig.LockDelayInMilliSeconds,
	}

	sessionId, writeMeta, err := c.client.CreateSession(sessionEntry, nil)
	if err != nil {
		return err
	}
	log.Debug().
		Str("SessionId", sessionId).
		Str("Duration", writeMeta.RequestTime.String()).
		Msg("consul session created")

	c.renewSessionPeriodically(sessionId, consulConfig.SessionRenewalTTLInSeconds, sessionRenewalChannel)
	c.SessionId = sessionId
	return err
}

func (c *ConsulClusterController) DestroySession() error {
	return c.client.DestroySession(c.SessionId)
}

func (c ConsulClusterController) GetLeader(serviceKey string) (cluster_types.Node, error) {
	kvPair, err := c.client.GetKVPair(serviceKey)
	nodeConfig := cluster_types.Node{}
	if err != nil {
		return nodeConfig, err
	}

	if kvPair == nil || kvPair.Session == "" {
		return nodeConfig, ErrLeaderNotFound
	}
	err = json.Unmarshal(kvPair.Value, &nodeConfig)
	return nodeConfig, err
}

func (c *ConsulClusterController) AcquireLock(node cluster_types.Node, serviceKey string) (bool, error) {
	lockData := node.Json()

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
			Msg("consul lock aquired on the session")
	}
	return acquired, err
}
