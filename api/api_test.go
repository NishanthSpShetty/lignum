package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NishanthSpShetty/lignum/config"
	"github.com/NishanthSpShetty/lignum/message"
	"github.com/stretchr/testify/assert"
)

const dummyServiceId = "DummyServiceId"

func TestGetMessage(t *testing.T) {
	requestData := GetMessageRequest{From: 0, To: 1}
	req, _ := json.Marshal(requestData)

	messageChannel := make(chan message.Message, 10)
	msg := message.New(config.Message{})
	server := NewServer(dummyServiceId, messageChannel, config.Server{}, msg)

	requestHandler := server.handleMessage()
	responseData := GetMessageResponse{}

	t.Run("returns empty message when no messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)

		expected := []message.Message{}
		assert.Equal(t, expected, responseData.Messages)
	})

	t.Run("returns expected message when messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)

		expected := []message.Message{}
		assert.Equal(t, expected, responseData.Messages)
	})
}
