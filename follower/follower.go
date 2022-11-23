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
	ready       bool
	messageStat []*cluster_types.MessageStat
}

type FollowerRegistry struct {
	follower map[string]*Follower
	queue    chan *Follower
}

func (f *Follower) IsHealthy() bool { return f.healthy }

//IsReady return true when follower is ready to recieve replicate message.
//currently we will  use healthy flag to mark as ready
func (f *Follower) IsReady() bool            { return f.ready }
func (f *Follower) Node() cluster_types.Node { return f.node }

//mark ready and healthy
func (f *Follower) MarkReady()   { f.ready = true }
func (f *Follower) MarkHealthy() { f.healthy = true }

//TopicOffset returns the latest offset, replicated message offset in
//follower node
func (f *Follower) TopicOffset(topic string) uint64 {
	for _, s := range f.messageStat {
		if s.Topic == topic {
			return s.Offset
		}
	}

	return 0
}

func (f *Follower) UpdateStat(topic string, offset uint64) {
	for _, s := range f.messageStat {
		if s.Topic == topic {
			s.Offset = offset
			return
		}
	}

	//means we are missing the topics
	//add to to stat
	stat := &cluster_types.MessageStat{
		Topic:  topic,
		Offset: offset,
	}
	f.messageStat = append(f.messageStat, stat)
}

func (f *FollowerRegistry) Register(fr cluster_types.FollowerRegistration) {
	//we know that the node is healthy when registering itself
	stats := make([]*cluster_types.MessageStat, 0, len(fr.MessageStat))
	for _, stat := range fr.MessageStat {
		stats = append(stats, &stat)
	}
	follower := &Follower{
		node:        fr.Node,
		healthy:     true,
		ready:       false,
		messageStat: stats,
	}

	f.follower[fr.Node.Id] = follower
	fmt.Printf("registered follower %v\n", follower)
	//add the follower data to queue too
	f.queue <- follower
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

func New(followerQueue chan *Follower) *FollowerRegistry {
	return &FollowerRegistry{
		follower: make(map[string]*Follower),
		queue:    followerQueue,
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
		//log.Debug().Int("healthy", healthy).Int("dead", dead).Msg("HealthStat")
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
				log.Debug().Msg("stopping follower health check service")
				ticker.Stop()
				return
			case <-ticker.C:
				f.healthCheck(client)
			}
		}
	}()
}
