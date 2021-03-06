package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMessage(t *testing.T) {
	t.Run("returns message from lignum", func(t *testing.T) {

		request, _ := http.NewRequest(http.MethodGet, "/api/message", nil)
		response := httptest.NewRecorder()

	})
}
