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
	InitialSizePerTopic int64 `yaml:"initial-size-per-topic"`
	//TODO: required for persistence
	//MessageDir          string `yaml:"message-dir"`
	//MessageFlushIntervalInMilliSeconds time.Duration `yaml:"message-flush-interval-in-ms"`
}

type Follower struct {
	HealthCheckIntervalInSecond time.Duration `yaml:"healthcheck-interval-in-seconds"`
	//per follower ping timeout
	HealthCheckTimeoutInMilliSeconds           time.Duration `yaml:"healthcheck-ping-timeout-in-ms"`
	RegistrationOrHealthCheckIntervalInSeconds time.Duration `yaml:"registration-healthcheck-interval-in-seconds"`
}

type Replication struct {
	InternalQueueSize int `yaml:"internal-queue-size"`
}

type Config struct {
	Server      Server      `yaml:"server"`
	Consul      Consul      `yaml:"consul"`
	Message     Message     `yaml:"message"`
	Follower    Follower    `yaml:"follower"`
	Replication Replication `yaml:"replication"`
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
