package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
	//Suggestion : sf13/viper Config solution
)

type Server struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ServiceKey string `yaml:"service-key"`
}

type Consul struct {
	Host                                 string        `yaml:"host"`
	Port                                 int           `yaml:"port"`
	ServiceName                          string        `yaml:"service-name"`
	SessionTTLInSeconds                  int           `yaml:"session-ttl-in-seconds"`
	SessionRenewalTTLInSeconds           int           `yaml:"session-renewal-ttl-in-seconds"`
	LockDelayInMilliSeconds              time.Duration `yaml:"lock-delay-in-ms"`
	LeaderElectionIntervalInMilliSeconds time.Duration `yaml:"leader-election-interval-in-ms"`
}

type Message struct {
	InitialSizePerTopic uint64 `yaml:"initial-size-per-topic"`
	DataDir             string `yaml:"data-dir"`
}

type Follower struct {
	HealthCheckIntervalInSecond time.Duration `yaml:"healthcheck-interval-in-seconds"`
	//per follower ping timeout
	HealthCheckTimeoutInMilliSeconds           time.Duration `yaml:"healthcheck-ping-timeout-in-ms"`
	RegistrationOrLeaderCheckIntervalInSeconds time.Duration `yaml:"registration-leader-check-interval-in-seconds"`
}

type Replication struct {
	InternalQueueSize           int           `yaml:"internal-queue-size"`
	ClientTimeoutInMilliSeconds time.Duration `yaml:"client-timeout-in-ms"`
	WALReplicationPort          int           `yaml:"wal-replication-port"`
	WalSyncIntervalInSec        time.Duration `yaml:"wal-sync-interval-in-sec"`
}

type Wal struct {
	QueueSize int `yaml:"queue-size"`
}

type Config struct {
	Server      Server      `yaml:"server"`
	Consul      Consul      `yaml:"consul"`
	Message     Message     `yaml:"message"`
	Follower    Follower    `yaml:"follower"`
	Replication Replication `yaml:"replication"`
	Wal         Wal         `yaml:"wal"`
}

func GetConfig(fileName string) (Config, error) {
	var config Config
	file, err := os.Open(fileName)
	if err != nil {
		return Config{}, err
	}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err

}
