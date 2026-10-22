package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	router := NewRouter(nil)
	for _, tt := range []struct {
		name, method, origin, requestedMethod, headers string
		status                                         int
	}{
		{"allowed origin", "GET", allowedOrigin, "", "", 200},
		{"no origin", "GET", "", "", "", 200},
		{"other origin", "GET", "https://example.com", "", "", 403},
		{"preflight", "OPTIONS", allowedOrigin, "POST", "content-type, Authorization", 204},
		{"rejected preflight origin", "OPTIONS", "https://example.com", "POST", "content-type", 403},
		{"rejected method", "OPTIONS", allowedOrigin, "DELETE", "", 403},
		{"rejected header", "OPTIONS", allowedOrigin, "POST", "X-Unknown", 403},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := "/health"
			if tt.method == "OPTIONS" {
				path = "/person"
			}
			req := httptest.NewRequest(tt.method, path, nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", tt.requestedMethod)
			req.Header.Set("Access-Control-Request-Headers", tt.headers)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tt.status {
				t.Fatalf("status = %d, want %d", res.Code, tt.status)
			}
			wantOrigin := ""
			if tt.origin == allowedOrigin {
				wantOrigin = allowedOrigin
			}
			if got := res.Header().Get("Access-Control-Allow-Origin"); got != wantOrigin {
				t.Errorf("allow origin = %q, want %q", got, wantOrigin)
			}
			if tt.status == 204 {
				if got := res.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
					t.Errorf("allow methods = %q", got)
				}
				if got := res.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
					t.Errorf("allow headers = %q", got)
				}
			}
		})
	}
}
