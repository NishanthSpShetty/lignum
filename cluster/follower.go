package cluster

type follower struct {
	follower map[string]Node
}

func (f *follower) add(n Node) {
	f.follower[n.Id] = n
}

func (f *follower) list() []Node {
	l := make([]Node, 0)
	for _, node := range f.follower {
		l = append(l, node)
	}
	return l
}

//TODO:temporary hack, should have some form cluster init which does this
var _follower = follower{
	follower: make(map[string]Node),
}

func AddFollower(node Node) {
	_follower.add(node)
}

func GetFollowers() []Node {
	return _follower.list()
}
