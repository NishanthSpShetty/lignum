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
	"github.com/NishanthSpShetty/lignum/wal"
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
	walChannel := make(chan wal.Payload, 10)
	msg := message.New(config.Message{InitialSizePerTopic: 10}, walChannel)

	server, err := NewServer(dummyServiceId, messageChannel, config.Server{}, msg, follower.New(make(chan *follower.Follower)))
	assert.Nil(t, err)

	requestHandler := server.handleGet
	responseData := GetMessageResponse{}

	t.Run("returns error when no messages are written", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code, "invalid request returns error")
	})

	t.Run("returns expected message when messages are written", func(t *testing.T) {
		dummyMsg := []byte("this is dummy message")

		msg.Put(context.Background(), DummyTopic, dummyMsg)
		request, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(req))
		response := httptest.NewRecorder()
		requestHandler(response, request)
		json.Unmarshal(response.Body.Bytes(), &responseData)

		assert.Equal(t, string(dummyMsg), responseData.Messages[0].Data)
		assert.Equal(t, uint64(0), responseData.Messages[0].Id)
	})
}

func TestServerRouting(t *testing.T) {
	messageChannel := make(chan replication.Payload, 10)
	walChannel := make(chan wal.Payload, 10)
	msg := message.New(config.Message{InitialSizePerTopic: 10}, walChannel)

	server, err := NewServer(dummyServiceId, messageChannel, config.Server{}, msg, follower.New(make(chan *follower.Follower)))
	assert.Nil(t, err)

	t.Run("GET /ping returns 200 PONG", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "PONG", rec.Body.String())
	})

	t.Run("POST /ping returns 405 Method Not Allowed", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/ping", nil)
		rec := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("POST and GET /api/message through ServeMux", func(t *testing.T) {
		postBody, _ := json.Marshal(PutMessageRequest{Topic: "routed_topic", Message: "hello via mux"})
		postReq, _ := http.NewRequest(http.MethodPost, "/api/message", bytes.NewReader(postBody))
		postRec := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(postRec, postReq)

		assert.Equal(t, http.StatusOK, postRec.Code)
		assert.Equal(t, "application/json", postRec.Header().Get("Content-Type"))

		getReqBody, _ := json.Marshal(GetMessageRequest{Topic: "routed_topic", From: 0, To: 1})
		getReq, _ := http.NewRequest(http.MethodGet, "/api/message", bytes.NewReader(getReqBody))
		getRec := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(getRec, getReq)

		assert.Equal(t, http.StatusOK, getRec.Code)
		assert.Equal(t, "application/json", getRec.Header().Get("Content-Type"))

		var resp GetMessageResponse
		err := json.Unmarshal(getRec.Body.Bytes(), &resp)
		assert.Nil(t, err)
		assert.Equal(t, 1, resp.Count)
		assert.Equal(t, "hello via mux", resp.Messages[0].Data)
	})

	t.Run("Multiple NewServer instances do not conflict on mux", func(t *testing.T) {
		server2, err := NewServer("second-server", messageChannel, config.Server{}, msg, follower.New(make(chan *follower.Follower)))
		assert.Nil(t, err)
		assert.NotNil(t, server2)
	})
}
