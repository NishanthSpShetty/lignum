package cluster

import (
	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/message/types"
)

//State contains the state of the cluster node
type State struct {
	leader            bool
	connectedToLeader bool
	leaderNode        cluster_types.Node
}

//state It is created during the app startup
var state *State = &State{}

func (s *State) isLeader() bool {
	return state.leader
}

//markLeader mark this node as leader
func (s *State) markLeader() {
	state.leader = true
}

//unmarkLeader marks this node as not a leader
//unlikely sitautaion
func (s *State) unmarkLeader() {
	state.leader = false
}

//isConnectedLeader return true if this node is connected to leader
func (s *State) isConnectedLeader() bool {
	return s.connectedToLeader
}

func (s *State) setConnectedToLeader(yes bool) {
	s.connectedToLeader = yes
}

func (s *State) setLeaderNode(leaderNode *types.Node) {
	s.leaderNode = leaderNode
}

func (s *State) getLeader() *types.Node {
	return s.leaderNode
}
