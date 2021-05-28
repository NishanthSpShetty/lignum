package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NishanthSpShetty/lignum/cluster"
	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/rs/zerolog/log"
)

type Server struct {
	serviceId        string
	replicationQueue chan<- message.MessageT
	config           config.Server
	httpServer       http.Server
}

func (s *Server) Stop(ctx context.Context) {
	s.httpServer.Shutdown(ctx)
}

func NewServer(serviceId string, queue chan<- message.MessageT, config config.Server) *Server {

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return &Server{
		serviceId:        serviceId,
		config:           config,
		replicationQueue: queue,
		httpServer:       http.Server{Addr: address},
	}
}

func (s *Server) registerFollower() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestBody, _ := ioutil.ReadAll(req.Body)
		log.Info().Bytes("RequestBody", requestBody).Msg("Request received for follower registration")
		node := cluster.Node{}
		json.Unmarshal(requestBody, &node)
		//TODO: this is accessing some global state, looks odd between the flow
		cluster.AddFollower(node)
		fmt.Fprintf(w, "Follower registered. Node : [ %v ]\n", node)
	}
}

func (a *Server) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "PONG")
	}
}

//handleMessage dispatch to particular handlers based on request method
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
	http.HandleFunc("/api/message", s.handleMessage())
	return s.httpServer.ListenAndServe()
}
