package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	router := setupRoutes()

	tests := []struct {
		name          string
		method        string
		path          string
		expectedCode  int
		expectedTitle string
		expectProblem bool
	}{
		{
			name:          "OK",
			method:        http.MethodGet,
			path:          "/status/200",
			expectedCode:  http.StatusOK,
			expectedTitle: "OK",
		},
		{
			name:          "Accepted",
			method:        http.MethodGet,
			path:          "/status/202",
			expectedCode:  http.StatusAccepted,
			expectedTitle: "Accepted",
		},
		{
			name:          "Not Found",
			method:        http.MethodGet,
			path:          "/status/404",
			expectedCode:  http.StatusNotFound,
			expectedTitle: "Not Found",
			expectProblem: true,
		},
		{
			name:          "Internal Server Error",
			method:        http.MethodGet,
			path:          "/status/500",
			expectedCode:  http.StatusInternalServerError,
			expectedTitle: "Internal Server Error",
			expectProblem: true,
		},
		{
			name:          "Unknown status code",
			method:        http.MethodGet,
			path:          "/status/598",
			expectedCode:  598,
			expectedTitle: "Unknown Status",
			expectProblem: true,
		},
		{
			name:          "Non-numeric code",
			method:        http.MethodGet,
			path:          "/status/invalid",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
		{
			name:          "Code below range",
			method:        http.MethodGet,
			path:          "/status/99",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
		{
			name:          "Informational code",
			method:        http.MethodGet,
			path:          "/status/100",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
		{
			name:          "Code above range",
			method:        http.MethodGet,
			path:          "/status/600",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
		{
			name:          "Deprecated status",
			method:        http.MethodGet,
			path:          "/status/305",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
		{
			name:          "Unused status",
			method:        http.MethodGet,
			path:          "/status/306",
			expectedCode:  http.StatusBadRequest,
			expectedTitle: "Bad Request",
			expectProblem: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Fatalf(
					"expected HTTP status %d, got %d",
					tt.expectedCode,
					rec.Code,
				)
			}

			if tt.expectProblem {
				var resp ProblemResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode problem response: %v", err)
				}

				if resp.Status != tt.expectedCode {
					t.Errorf("expected problem status %d, got %d", tt.expectedCode, resp.Status)
				}

				if resp.Title != tt.expectedTitle {
					t.Errorf("expected title %q, got %q", tt.expectedTitle, resp.Title)
				}

				if resp.Type != "about:blank" {
					t.Errorf("expected problem type about:blank, got %q", resp.Type)
				}

				if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
					t.Errorf("expected application/problem+json, got %q", got)
				}

				return
			}

			var resp StatusResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode status response: %v", err)
			}

			if resp.Code != tt.expectedCode {
				t.Errorf("expected body code %d, got %d", tt.expectedCode, resp.Code)
			}

			if resp.Status != tt.expectedTitle {
				t.Errorf("expected status %q, got %q", tt.expectedTitle, resp.Status)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("expected application/json, got %q", got)
			}
		})
	}
}

func TestStatusHeadResponseHasNoBody(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodHead, "/status/404", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("HEAD response should have no body, got %q", rec.Body.String())
	}

	if rec.Header().Get("Content-Length") == "" {
		t.Error("expected Content-Length for the corresponding JSON representation")
	}
}

func TestMultipleChoicesResponse(t *testing.T) {
	router := setupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/status/300", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultipleChoices {
		t.Fatalf("expected status 300, got %d", rec.Code)
	}

	var resp StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Alternatives) != 2 {
		t.Fatalf("expected two alternatives, got %v", resp.Alternatives)
	}

	if resp.Alternatives[0] != "/health" || resp.Alternatives[1] != "/version" {
		t.Errorf("unexpected alternatives: %v", resp.Alternatives)
	}
}
