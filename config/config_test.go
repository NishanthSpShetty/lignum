package config

import (
	"testing"
)

func Test_ReadConfig(t *testing.T) {

	//working directory would be same as this package directory while running the test
	configFile := "../test_data/config_test.yml"
	config, err := GetConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if config.Server.Host != "localhost" {
		t.Fatalf("Expected %s, Got %s.",
			"localhost", config.Consul.Host)
	}
}
