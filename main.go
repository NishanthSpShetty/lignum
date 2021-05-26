package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initialiseLogger(development bool) {
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	if development {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	initialiseLogger(true)

	configFile := flag.String("config", "", "get configuration from file")
	flag.Parse()

	appConfig, err := config.GetConfig(*configFile)
	if err != nil {
		//o	 ("Failed to read config", err)
		log.Error().Err(err).Msg("Failed to read config")
		return
	}

	log.Debug().Interface("Config", appConfig).Msg("Loaded app config")
	service, err := service.New(appConfig)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialise service")
		return
	}
	err = service.Start()

	if err != nil {
		fmt.Println(err)
		log.Error().Err(err).Msg("failed to start service")
		return
	}

}
