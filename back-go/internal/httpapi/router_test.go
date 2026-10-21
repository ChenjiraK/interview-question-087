package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	tests := []struct {
		name   string
		path   string
		status int
		key    string
		value  string
	}{
		{"health", "/health", http.StatusOK, "status", "ok"},
		{"ping", "/api/v1/ping", http.StatusOK, "message", "pong"},
		{"unknown route", "/missing", http.StatusNotFound, "error", "route not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want JSON", got)
			}

			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body[tt.key] != tt.value {
				t.Errorf("body[%q] = %q, want %q", tt.key, body[tt.key], tt.value)
			}
		})
	}
}
