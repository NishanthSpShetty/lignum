package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/metrics"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/rs/zerolog/log"
)

// request message struct
type PutMessageRequest struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type GetMessageRequest struct {
	//will need range to pick the messages from
	Topic string `json:"topic"`
	From  uint64 `json:"from"`
	To    uint64 `json:"to"`
}

//respons message struct
type GetMessageResponse struct {
	Messages []message.Message `json:"messages"`
	Count    int               `json:"count"`
}

func (s *Server) handlePost(w http.ResponseWriter, req *http.Request) {
	metrics.IncrementPostRequest()
	var msg PutMessageRequest

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	ctx := req.Context()

	err := decoder.Decode(&msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to read request body %s ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if msg.Topic == "" {
		http.Error(w, "message topic not specified", http.StatusBadRequest)
		log.Error().Msg("message topic not specified")
		return
	}

	log.Debug().Str("Data", msg.Message).Str("Topic", msg.Topic).Msg("message received")
	mesg := s.message.Put(ctx, msg.Topic, msg.Message)
	//write messages to replication queue
	payload := replication.Payload{
		Topic: msg.Topic,
		Id:    mesg.Id,
		Data:  mesg.Data,
	}
	s.replicationQueue <- payload

	fmt.Fprintf(w, "{\"status\": \"message commited\", \"data\": \"%s\"}", msg.Message)
}

func (s *Server) handleGet(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	var messageRequest GetMessageRequest = GetMessageRequest{}
	err := decoder.Decode(&messageRequest)

	if err != nil {
		log.Error().Err(err).Msg("failed to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	from := messageRequest.From
	to := messageRequest.To

	if from < 0 || to <= from {
		http.Error(w, "invalid message range (must: from<to)", http.StatusBadRequest)
		log.Error().Uint64("From", from).Uint64("To", to).Msg("invalid range specified")
		return
	}

	if messageRequest.Topic == "" {
		http.Error(w, "message topic not specified", http.StatusBadRequest)
		log.Error().Msg("message topic not specified")
		return

	}

	if !s.message.TopicExist(messageRequest.Topic) {
		http.Error(w, "topic does not exist", http.StatusBadRequest)
		log.Error().Str("Topic", messageRequest.Topic).Msg("topic does not exist")
		return
	}

	messages := s.message.Get(messageRequest.Topic, from, to)
	messag := GetMessageResponse{Messages: messages, Count: len(messages)}

	log.Debug().Interface("ReceivedMessage", messageRequest).Send()
	json.NewEncoder(w).Encode(messag)
}
