package types

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/rs/zerolog/log"
)

// Leader has port information of the leader
type Leader struct {
	Port int
}

type ClusterController interface {
	CreateSession(config.Consul, chan struct{}) error
	AcquireLock(Node, string) (bool, error)
	GetLeader(string) (Node, error)
	DestroySession() error
}

// NodeNodeConfig contains the node information
type Node struct {
	Id              string `json:"id"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ReplicationPort int    `json:"replication_port"`
	_json           []byte
}

func (n Node) getJson() ([]byte, error) {
	return json.Marshal(n)
}

func (n Node) Json() []byte {
	if n._json == nil {
		n._json, _ = json.Marshal(n)
	}
	return n._json
}

func NewNode(id string, host string, port, rp int) Node {
	return Node{
		Id:              id,
		Host:            host,
		Port:            port,
		ReplicationPort: rp,
	}
}

type MessageStat struct {
	Topic  string `json:"topic"`
	Offset uint64 `json:"offset"`
}

// FollowerRegistration follower registrartion request data
type FollowerRegistration struct {
	Node        Node          `json:"node"`
	MessageStat []MessageStat `json:"message-stat"`
}

func (n *Node) Ping(client http.Client) bool {
	pingUrl := fmt.Sprintf("http://%s:%d/ping", n.Host, n.Port)
	response, err := client.Get(pingUrl)
	if err != nil {
		log.Error().RawJSON("node", n.Json()).Err(err).Msg("ping failed")
		return false
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		return true
	}
	// anything else return false, not expecting any other value apart from status OK(200)
	return false
}
