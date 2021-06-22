package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/follower"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/NishanthSpShetty/lignum/replication"
	"github.com/stretchr/testify/assert"
)

const (
	dummyServiceId = "DummyServiceId"
	DummyTopic     = "DummyTopic"
)

func TestGetMessage(t *testing.T) {
	requestData := GetMessageRequest{Topic: DummyTopic, From: 0, To: 1}
	req, _ := json.Marshal(requestData)

	messageChannel := make(chan replication.Payload, 10)
	msg := message.New(config.Message{})

	server := NewServer(dummyServiceId, messageChannel, config.Server{}, msg, follower.New())

	requestHandler := server.handleMessage()
	responseData := GetMessageResponse{}

	t.Run("returns error when no messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code, "invalid request returns error")

	})

	t.Run("returns expected message when messages are written", func(t *testing.T) {
		dummyMsg := "this is dummy message"

		msg.Put(context.Background(), DummyTopic, dummyMsg)
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)

		expected := []message.Message{{Id: 0, Data: dummyMsg}}
		assert.Equal(t, expected, responseData.Messages)
	})
}
