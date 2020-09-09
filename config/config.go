package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	ServiceKey string `yaml:"service-key"`
}

type Consul struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	ServiceName       string `yaml:"service-name"`
	SessionTTL        string `yaml:"sessionTTL"`
	SessionRenewalTTL string `yaml:"sessionRenewalTTL"`
}

type Config struct {
	Server Server `yaml:"server"`
	Consul Consul `yaml:"consul"`
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