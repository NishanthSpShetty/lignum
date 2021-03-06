package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lignum/message"
	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	requestData := GetMessageRequest{Key: "foo"}
	req, _ := json.Marshal(requestData)
	messageChannel := make(chan message.MessageT)
	requestHandler := handleMessage(messageChannel)
	responseData := GetMessageResponse{}

	t.Run("returns empty message when no messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)
		assert.Equal(t, "", responseData.Value)
	})

	t.Run("returns expected message when messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)
		assert.Equal(t, "bar", responseData.Value)
	})
}
