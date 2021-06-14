package follower

import (
	"context"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster"
)

type Follower struct {
	node    cluster.Node
	healthy bool
}

type FollowerRegistry struct {
	follower map[string]*Follower
}

func (f *FollowerRegistry) Register(n cluster.Node) {
	//we know that the node is healthy when registering itself
	f.follower[n.Id] = &Follower{node: n, healthy: true}
}

func (f *FollowerRegistry) List() []cluster.Node {
	l := make([]cluster.Node, 0)
	for _, follower := range f.follower {
		l = append(l, follower.node)
	}
	return l
}

func New() *FollowerRegistry {
	return &FollowerRegistry{
		follower: make(map[string]*Follower),
	}
}

func (f *FollowerRegistry) healthCheck() {
}

func (f *FollowerRegistry) StartHealthCheck(ctx context.Context, healthCheckFrequency time.Duration) {
	go func() {
		ticker := time.NewTicker(healthCheckFrequency)

		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			f.healthCheck()
		}
	}()
}
