package config

import (
	"os"
	//Suggestion : sf13/viper Config solution
	"gopkg.in/yaml.v2"
)

type Server struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ServiceKey string `yaml:"service-key"`
	ServiceId  string
}

type Consul struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	ServiceName       string `yaml:"service-name"`
	SessionTTL        string `yaml:"sessionTTL"`
	SessionRenewalTTL string `yaml:"sessionRenewalTTL"`
}

type Message struct {
	InitialSize int64 `yaml:"initial-size"`
}

type Config struct {
	Server  Server  `yaml:"server"`
	Consul  Consul  `yaml:"consul"`
	Message Message `yaml:"message"`
}

func (c *Config) SetServiceId(id string) {
	c.Server.ServiceId = id
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
