package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expectedBody := "OK\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestHealthzHandler(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expectedBody := "OK\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestVersionHandler(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Version != Version {
		t.Errorf("expected version %q, got %q", Version, resp.Version)
	}
}

func TestEchoHandler(t *testing.T) {
	router := setupRoutes()

	t.Run("GET with query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/echo?key=value&foo=bar", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp EchoResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if resp.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", resp.Method)
		}

		if len(resp.Query["key"]) == 0 || resp.Query["key"][0] != "value" {
			t.Errorf("expected query param key=value, got %v", resp.Query["key"])
		}

		if len(resp.Query["foo"]) == 0 || resp.Query["foo"][0] != "bar" {
			t.Errorf("expected query param foo=bar, got %v", resp.Query["foo"])
		}
	})

	t.Run("POST with JSON body and custom header", func(t *testing.T) {
		jsonBody := `{"message":"hello world"}`
		req := httptest.NewRequest(
			http.MethodPost,
			"/echo",
			bytes.NewBufferString(jsonBody),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Custom-Header", "fakeway-test")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp EchoResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if resp.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", resp.Method)
		}

		if resp.Body != jsonBody {
			t.Errorf("expected body %q, got %q", jsonBody, resp.Body)
		}

		if len(resp.Headers["X-Custom-Header"]) == 0 ||
			resp.Headers["X-Custom-Header"][0] != "fakeway-test" {
			t.Errorf(
				"expected X-Custom-Header 'fakeway-test', got %v",
				resp.Headers["X-Custom-Header"],
			)
		}
	})
}
