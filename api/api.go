package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/NishanthSpShetty/lignum/cluster/types"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	serviceId        string
	replicationQueue chan<- replication.Payload
	config           config.Server
	httpServer       *http.Server
	message          *message.MessageStore
	follower         *follower.FollowerRegistry
	listener         net.Listener
}

func (s *Server) Stop(ctx context.Context) {
	s.httpServer.Shutdown(ctx)
}

func NewServer(serviceId string, queue chan<- replication.Payload, config config.Server, message *message.MessageStore, follower *follower.FollowerRegistry) (*Server, error) {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	httpServer := http.Server{
		Addr: address,
		// shamelessly copied following config from internet, will revisit this later
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "NewServer failed to listen on given address ")
	}

	return &Server{
		serviceId:        serviceId,
		config:           config,
		replicationQueue: queue,
		httpServer:       &httpServer,
		message:          message,
		follower:         follower,
		listener:         ln,
	}, nil
}

func (s *Server) registerFollower() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Info().Bytes("RequestBody", requestBody).Msg("request received for follower registration")
		fr := types.FollowerRegistration{}
		json.Unmarshal(requestBody, &fr)
		s.follower.Register(fr)
		fmt.Fprintf(w, "Follower registered. Node : [ %v ]\n", fr)
	}
}

func (a *Server) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "PONG")
	}
}

// handleMessage dispatch to particular handlers based on request method
func (s *Server) handleMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "POST":
			s.handlePost(w, req)
		case "GET":
			s.handleGet(w, req)
		default:
			http.Error(w, "request method must be one of [ GET, POST ].", http.StatusMethodNotAllowed)
		}
	}
}

func (s *Server) Serve() error {
	log.Info().
		Str("Host", s.config.Host).
		Int("Port", s.config.Port).
		Msg("Starting HTTP service")

	http.HandleFunc("/ping", s.ping())
	http.HandleFunc("/api/follower/register", s.registerFollower())
	http.HandleFunc("/internal/api/replicate", s.replicate())
	http.HandleFunc("/api/message", s.handleMessage())
	http.HandleFunc("/api/topics", s.TopicHandler())
	http.Handle("/metrics", promhttp.Handler())
	return s.httpServer.Serve(s.listener)
}
