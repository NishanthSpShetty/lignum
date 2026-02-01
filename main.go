package main

import (
	"flag"
	"os"
	"strings"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	ENV_DEVELOPMENT = "DEVELOPMENT"
)

func initialiseLogger(development bool, level zerolog.Level) {
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()
	if development {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(level)
}

func isDevelopment(dev_mode string) bool {
	return strings.ToUpper(strings.TrimSpace(dev_mode)) == ENV_DEVELOPMENT
}

func main() {
	isDev := false
	if envName, ok := os.LookupEnv("ENV"); ok && isDevelopment(envName) {
		isDev = true
	}

	envLogLevel, present := os.LookupEnv("LOG_LEVEL")
	logLevel := zerolog.DebugLevel
	var err error
	if present {
		logLevel, err = zerolog.ParseLevel(envLogLevel)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse log level")
			return
		}
	}

	initialiseLogger(isDev, logLevel)

	configFile := flag.String("config", "", "get configuration from file")
	flag.Parse()

	appConfig, err := config.GetConfig(*configFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to read config")
		return
	}

	log.Debug().Interface("config", appConfig).Msg("loaded app config")
	service, err := service.New(appConfig)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialise service")
		return
	}
	err = service.Start()

	if err != nil {
		log.Error().Err(err).Send()
		if !service.Stopped() {
			service.Stop()
		}
		return
	}
}
