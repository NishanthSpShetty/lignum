package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/proto"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Server struct {
	serviceId        string
	replicationQueue chan<- replication.Payload
	config           config.Server
	httpServer       *http.Server
	grpcServer       *grpc.Server
	message          *message.MessageStore
	follower         *follower.FollowerRegistry
	httpListener     net.Listener
	proto.UnimplementedLignumServer
}

func (s *Server) Stop(ctx context.Context) {
	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("error shutting down http server")
	}
}

func NewServer(serviceId string, queue chan<- replication.Payload, config config.Server, message *message.MessageStore, follower *follower.FollowerRegistry) (*Server, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "NewServer failed to listen on given address")
	}

	s := &Server{
		serviceId:        serviceId,
		config:           config,
		replicationQueue: queue,
		message:          message,
		follower:         follower,
		httpListener:     ln,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	s.setupGrpc()
	return s, nil
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", s.handlePing)
	mux.HandleFunc("POST /api/follower/register", s.handleRegisterFollower)
	mux.HandleFunc("POST /internal/api/replicate", s.handleReplicate)
	mux.HandleFunc("GET /api/message", s.handleGet)
	mux.HandleFunc("POST /api/message", s.handlePost)
	mux.HandleFunc("GET /api/topics", s.handleTopicGet)
	mux.Handle("GET /metrics", promhttp.Handler())
}

func (s *Server) handleRegisterFollower(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, 1024*1024)
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read request body")
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	log.Info().Bytes("request_body", requestBody).Msg("request received for follower registration")
	fr := types.FollowerRegistration{}
	if err := json.Unmarshal(requestBody, &fr); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	s.follower.Register(fr)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Follower registered. Node : [ %v ]\n", fr)
}

func (s *Server) handlePing(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "PONG")
}

func (s *Server) Start() error {
	log.Info().
		Str("Host", s.config.Host).
		Int("Port", s.config.Port).
		Msg("Starting HTTP service")

	go s.startGrpc()
	return s.httpServer.Serve(s.httpListener)
}
