package clusetr

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

var client *api.Client

func InitialiseConsulClient(host string, port int) error {

	var err error

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", host, port)

	client, err = api.NewClient(config)

	return err
}

func CreateConsulSession(serviceName, ttlS string) (string, error) {

	sessionEntry := &api.SessionEntry{
		Name:      serviceName,
		TTL:       ttlS,
		LockDelay: 1 * time.Millisecond,
	}

	sessionId, queryDuration, err := client.Session().Create(sessionEntry, nil)

	log.Debugf("Consul session created, ID : %v, Aquired in :%dms  ", sessionId, queryDuration.RequestTime.Milliseconds())

	return sessionId, err
}
