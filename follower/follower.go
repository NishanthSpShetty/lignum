package follower

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/rs/zerolog/log"
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
	fmt.Println("registered")
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

func isActive(client http.Client, node cluster.Node) bool {
	pingUrl := fmt.Sprintf("http://%s:%d/ping", node.Host, node.Port)
	response, err := client.Get(pingUrl)
	if err != nil {
		log.Error().Err(err).Msg("ping failed")
		return false
	}

	if response.StatusCode == http.StatusOK {
		return true
	}
	//anything else return false, not expecting any other value apart from status OK(200)
	return false
}

func (f *FollowerRegistry) healthCheck(client http.Client) {
	//iterate over each follower nodes and mark them healthy
	healthy := 0
	dead := 0
	for _, follower := range f.follower {
		//if we marked node as unhealthy, dont check again.
		//TODO: have some multiple tries before considering the node as dead.
		//timeouts can happen even when the node is healthy
		if !follower.healthy || !isActive(client, follower.node) {
			//mark the follower as not healthy,
			//TODO: remove the dead followers in cleanup
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

func (f *FollowerRegistry) StartHealthCheck(ctx context.Context, healthCheckFrequency time.Duration) {
	//create http client
	client := http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Timeout: 5 * time.Millisecond,
	}
	go func() {
		log.Debug().Msg("starting healthCheck service")
		ticker := time.NewTicker(healthCheckFrequency)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("cancle healthCheck")
				ticker.Stop()
				return
			case <-ticker.C:
				f.healthCheck(client)
			}
		}
	}()
}
