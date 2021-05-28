package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func (s *Service) Stop() {
	log.Info().
		Str("ServiceID", s.ServiceId).
		Msg("Stopping all routines, channels")

	//call all cancelfunctions
	for _, cancel := range s.Cancels {
		cancel()
	}
	//close all channels
	close(s.SessionRenewalChannel)
	close(s.ReplicationQueue)

	err := s.ClusterController.DestroySession()

	if err != nil {
		log.Error().Err(err).Msg("Failed to destroy the session ")
	}
	s.apiServer.Stop(context.Background())
	log.Info().Str("ServiceId", s.ServiceId).Msg("Shutting down")
}

func (s *Service) signalHandler() {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
		//read the signal and discard
		<-signalChannel
		s.Stop()
		os.Exit(0)
	}()
}
