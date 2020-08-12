package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
)

func main() {
	fmt.Println("hello")
	log.Info("Starting producer service")
	api.DefaultConfig()
}
