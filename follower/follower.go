package follower

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cluster_types "github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/rs/zerolog/log"
)

type Follower struct {
	node    cluster_types.Node
	healthy bool
	//ready for replication
	ready bool
}

type FollowerRegistry struct {
	follower map[string]*Follower
}

func (f *Follower) IsHealthy() bool { return f.healthy }

//IsReady return true when follower is ready to recieve replicate message.
//currently we will  use healthy flag to mark as ready
func (f *Follower) IsReady() bool            { return f.healthy }
func (f *Follower) Node() cluster_types.Node { return f.node }

func (f *FollowerRegistry) Register(n cluster_types.Node) {
	//we know that the node is healthy when registering itself
	//for now we will mark the follower node as replication ready node
	f.follower[n.Id] = &Follower{node: n, healthy: true, ready: true}
	fmt.Println("registered")
}

func (f *FollowerRegistry) ListNodes() []cluster_types.Node {
	l := make([]cluster_types.Node, 0)
	for _, follower := range f.follower {
		l = append(l, follower.node)
	}
	return l
}

func (f *FollowerRegistry) List() map[string]*Follower {
	return f.follower
}

func New() *FollowerRegistry {
	return &FollowerRegistry{
		follower: make(map[string]*Follower),
	}
}

func isActive(client http.Client, node *cluster_types.Node) bool {
	return node.Ping(client)
}

func (f *FollowerRegistry) healthCheck(client http.Client) {
	//iterate over each follower nodes and mark them healthy
	healthy := 0
	dead := 0
	for _, follower := range f.follower {
		//if we marked node as unhealthy, dont check again.
		//TODO: have some multiple tries before considering the node as dead.
		//timeouts can happen even when the node is healthy
		if !follower.healthy {
			dead += 1
			continue
		}
		if !isActive(client, &follower.node) {
			//mark the follower as not healthy,
			//TODO: remove the dead followers in cleanup process
			follower.healthy = false
			dead += 1
			continue
		}
		healthy += 1
	}

	if healthy|dead != 0 {
		log.Debug().Int("healthy", healthy).Int("dead", dead).Msg("HealthStat")
	}
}

func (f *FollowerRegistry) StartHealthCheck(ctx context.Context, healthCheckInterval time.Duration, healthCheckTimeout time.Duration) {
	//create http client
	client := http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		// 		Timeout: 5 * time.Millisecond,
		Timeout: healthCheckTimeout,
	}
	go func() {
		log.Debug().Msg("starting health check service")
		ticker := time.NewTicker(healthCheckInterval)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("stopping follwer health check service")
				ticker.Stop()
				return
			case <-ticker.C:
				f.healthCheck(client)
			}
		}
	}()
}
