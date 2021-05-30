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
	Message string
}

type GetMessageRequest struct {
	//will need range to pick the messages from
	From int
	To   int
}

//respons message struct
type GetMessageResponse struct {
	Messages []string `json:"messages"`
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
	messages := message.Get(messageRequest.From, messageRequest.To)
	messag := GetMessageResponse{Messages: messages}

	log.Debug().Interface("RecievedMessage", messageRequest).Send()
	json.NewEncoder(w).Encode(messag)
}
