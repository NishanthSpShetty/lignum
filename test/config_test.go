package test

import (
	"testing"

	"github.com/disquote-logger/config"
)

func Test_ReadConfig(t *testing.T) {

	configFile := "config_test.yml"
	config, err := config.GetConfig(configFile)
	if err != nil {
		t.Fatal(err)

	}
	if config.Server.Host != "localhost" {
		t.Fatalf("Expected %s, Got %s.",
			"localhost", config.Consul.Host)
	}
}
