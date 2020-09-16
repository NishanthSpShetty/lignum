package cluster

import (
	"fmt"
	"time"

	"github.com/disquote-logger/config"
	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

var client *api.Client

func InitialiseConsulClient(consulConfig config.Consul) error {

	var err error

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", consulConfig.Host, consulConfig.Port)

	client, err = api.NewClient(config)

	return err
}

func renewSessionPeriodicall(sessionId string, ttlS string, sessionRenewalChannel chan struct{}) {
	defer close(sessionRenewalChannel)
	for {
		client.Session().RenewPeriodic(ttlS, sessionId, nil, sessionRenewalChannel)
		time.Sleep(90 * time.Second)
	}
}

func CreateConsulSession(consulConfig config.Consul, sessionRenewalChannel chan struct{}) (string, error) {

	sessionEntry := &api.SessionEntry{
		Name:      consulConfig.ServiceName,
		TTL:       consulConfig.SessionTTL,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, queryDuration, err := client.Session().Create(sessionEntry, nil)

	if err != nil {
		return "", err
	}
	log.Debugf("Consul session created, ID : %v, Aquired in :%dms  ", sessionId, queryDuration.RequestTime.Milliseconds())
	//TODO : should it be started conditionally?
	go renewSessionPeriodicall(sessionId, consulConfig.SessionRenewalTTL, sessionRenewalChannel)
	return sessionId, err
}

func DestroySession(sessionId string) error {

	_, err := client.Session().Destroy(sessionId, nil)
	return err
}

func GetLeader(serviceKey string) (*api.KVPair, error) {
	kv, _, err := client.KV().Get(serviceKey, nil)
	return kv, err

}
