package cluster

import (
	"errors"
	"strconv"
)

var (
	errLeaderNotFound = errors.New("failed to get reader from the cluster controller")
)

//Leader has port information of the leader
type Leader struct {
	Port int
}

type ClusterController interface {
	AquireLock(NodeConfig, string) (bool, error)
	GetLeader(string) (Leader, error)
}

//NodeNodeConfig contains the node information
type NodeConfig struct {
	NodeId string
	NodeIp string
	Port   int
}

func (nodeConfig NodeConfig) Stringer() string {
	return "NodeConfig { NodeId : " + nodeConfig.NodeId + ", NodeIp : " + nodeConfig.NodeId + ", Port " + strconv.Itoa(nodeConfig.Port) + "}"
}

func NewNodeConfig(nodeId string, nodeIp string, port int) NodeConfig {
	return NodeConfig{
		NodeId: nodeId,
		NodeIp: nodeId,
		Port:   port,
	}
}
