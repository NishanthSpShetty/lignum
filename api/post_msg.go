package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NishanthSpShetty/lignum/message"
	"github.com/rs/zerolog/log"
)

// request message struct
type PutMessageRequest struct {
	Message string `json:"message"`
}

type GetMessageRequest struct {
	//will need range to pick the messages from
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

//respons message struct
type GetMessageResponse struct {
	Messages []message.Message `json:"messages"`
}

func (s *Server) handlePost(w http.ResponseWriter, req *http.Request) {
	var msg PutMessageRequest

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	ctx := req.Context()

	err := decoder.Decode(&msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body %s ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Debug().Str("RecievedMessage", msg.Message).Send()
	message.Put(ctx, msg.Message)

	fmt.Fprintf(w, "{status : \"message commited\"\n message : { %v }", "key:value")
}

func (s *Server) handleGet(w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	var messageRequest GetMessageRequest = GetMessageRequest{}
	err := decoder.Decode(&messageRequest)

	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	from := messageRequest.From
	to := messageRequest.To

	if from < 0 || to <= from {
		http.Error(w, "invalid messge range (must: from<to)", http.StatusBadRequest)
		log.Error().Uint64("From", from).Uint64("To", to).Msg("invalid range specified")
		return
	}
	messages := message.Get(from, to)
	messag := GetMessageResponse{Messages: messages}

	log.Debug().Interface("RecievedMessage", messageRequest).Send()
	json.NewEncoder(w).Encode(messag)
}
