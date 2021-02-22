package cluster

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lignum/config"
)

var (
	errLeaderNotFound = errors.New("failed to get leader from the cluster controller")
)

//Leader has port information of the leader
type Leader struct {
	Port int
}

type ClusterController interface {
	CreateSession(config.Consul, chan struct{}) error
	AquireLock(NodeConfig, string) (bool, time.Duration, error)
	GetLeader(string) (NodeConfig, error)
	DestroySession() error
}

//NodeNodeConfig contains the node information
type NodeConfig struct {
	NodeId string `json:"node-id"`
	NodeIp string `json:"node-ip"`
	Port   int    `json:"port"`
}

func (nodeConfig NodeConfig) json() ([]byte, error) {
	return json.Marshal(nodeConfig)
}

func NewNodeConfig(nodeId string, nodeIp string, port int) NodeConfig {
	return NodeConfig{
		NodeId: nodeId,
		NodeIp: nodeIp,
		Port:   port,
	}
}
