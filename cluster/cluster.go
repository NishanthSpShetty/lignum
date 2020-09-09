package cluster

import "strconv"

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
