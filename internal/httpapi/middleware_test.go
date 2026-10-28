package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodOptions, "/echo", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS preflight, got %d", rec.Code)
	}

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", origin)
	}

	if methods := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "HEAD") {
		t.Errorf("expected CORS methods to include HEAD, got %q", methods)
	}
}
