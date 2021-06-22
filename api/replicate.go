package api

import (
	"encoding/json"
	"net/http"

	"github.com/NishanthSpShetty/lignum/metrics"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/rs/zerolog/log"
)

func (s *Server) handleReplicate(w http.ResponseWriter, req *http.Request) {

	metrics.IncrementReplicationRequest()

	payload := replication.Payload{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	//ctx := req.Context()

	err := decoder.Decode(&payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to read request body %s ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) replicate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		switch req.Method {
		case "POST":
			s.handleReplicate(w, req)
		default:
			http.Error(w, "request method must be [ POST ].", http.StatusMethodNotAllowed)
		}
	}
}
