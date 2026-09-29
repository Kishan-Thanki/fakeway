package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(jsonBody))
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

		if len(resp.Headers["X-Custom-Header"]) == 0 || resp.Headers["X-Custom-Header"][0] != "fakeway-test" {
			t.Errorf("expected X-Custom-Header 'fakeway-test', got %v", resp.Headers["X-Custom-Header"])
		}
	})
}

func TestStatusHandler(t *testing.T) {
	router := setupRoutes()

	tests := []struct {
		name           string
		path           string
		expectedCode   int
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "Valid status 200 OK",
			path:           "/status/200",
			expectedCode:   http.StatusOK,
			expectedStatus: "OK",
			expectError:    false,
		},
		{
			name:           "Valid status 404 Not Found",
			path:           "/status/404",
			expectedCode:   http.StatusNotFound,
			expectedStatus: "Not Found",
			expectError:    false,
		},
		{
			name:           "Valid status 500 Internal Server Error",
			path:           "/status/500",
			expectedCode:   http.StatusInternalServerError,
			expectedStatus: "Internal Server Error",
			expectError:    false,
		},
		{
			name:           "Unassigned status code within 100-599 range",
			path:           "/status/598",
			expectedCode:   598,
			expectedStatus: "Unknown Status",
			expectError:    false,
		},
		{
			name:         "Non-numeric code parameter",
			path:         "/status/invalid",
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "Status code below 100",
			path:         "/status/99",
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "Status code above 599",
			path:         "/status/600",
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected HTTP status code %d, got %d", tt.expectedCode, rec.Code)
			}

			if tt.expectError {
				var errResp map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error JSON: %v", err)
				}
				if _, ok := errResp["error"]; !ok {
					t.Errorf("expected response to contain 'error' key, got %v", errResp)
				}
			} else {
				var resp StatusResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode JSON response: %v", err)
				}
				if resp.Code != tt.expectedCode {
					t.Errorf("expected body code %d, got %d", tt.expectedCode, resp.Code)
				}
				if resp.Status != tt.expectedStatus {
					t.Errorf("expected body status %q, got %q", tt.expectedStatus, resp.Status)
				}
			}
		})
	}
}

func TestDelayHandler(t *testing.T) {
	router := setupRoutes()

	t.Run("Valid delay 10ms", func(t *testing.T) {
		start := time.Now()
		req := httptest.NewRequest(http.MethodGet, "/delay/10ms", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		if time.Since(start) < 10*time.Millisecond {
			t.Errorf("handler responded faster than the requested 10ms delay")
		}

		var resp DelayResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if resp.Duration != "10ms" {
			t.Errorf("expected duration '10ms', got %s", resp.Duration)
		}
	})

	t.Run("Invalid duration format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/delay/invalid", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Negative duration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/delay/-500ms", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Exceeds max allowed delay (30s)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/delay/35s", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("Context cancellation on client disconnect", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/delay/2s", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		router.ServeHTTP(rec, req)
		elapsed := time.Since(start)

		if elapsed >= 1*time.Second {
			t.Errorf("expected handler to cancel early, but took %v", elapsed)
		}
	})
}

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
}
