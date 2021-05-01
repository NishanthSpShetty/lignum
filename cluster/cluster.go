package cluster

import (
	"encoding/json"
	"time"

	"github.com/NishanthSpShetty/lignum/config"
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
	Id   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (nodeConfig Node) Json() ([]byte, error) {
	return json.Marshal(nodeConfig)
}

func NewNode(id string, host string, port int) Node {
	return Node{
		Id:   id,
		Host: host,
		Port: port,
	}
}
