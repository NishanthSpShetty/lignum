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
	AquireLock(Node, string) (bool, time.Duration, error)
	GetLeader(string) (Node, error)
	DestroySession() error
}

//NodeNodeConfig contains the node information
type Node struct {
	NodeId string `json:"node-id"`
	NodeIp string `json:"node-ip"`
	Port   int    `json:"port"`
}

func (nodeConfig Node) json() ([]byte, error) {
	return json.Marshal(nodeConfig)
}

func NewNodeConfig(nodeId string, nodeIp string, port int) Node {
	return Node{
		NodeId: nodeId,
		NodeIp: nodeIp,
		Port:   port,
	}
}
