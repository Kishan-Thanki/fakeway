package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
		defer cancel()

		req := httptest.NewRequest(
			http.MethodGet,
			"/delay/2s",
			nil,
		).WithContext(ctx)

		rec := httptest.NewRecorder()

		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		router.ServeHTTP(rec, req)
		elapsed := time.Since(start)

		if elapsed >= time.Second {
			t.Errorf("expected handler to cancel early, but took %v", elapsed)
		}
	})
}
