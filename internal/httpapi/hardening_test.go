package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strings"
	"testing"
)

func TestEchoBodyLimit(t *testing.T) {
	router := setupRoutes()

	t.Run("accepts body at limit", func(t *testing.T) {
		body := strings.Repeat("x", MAX_ECHO_BODY_SIZE)
		req := httptest.NewRequest(
			http.MethodPost,
			"/echo",
			strings.NewReader(body),
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("rejects body above limit", func(t *testing.T) {
		body := strings.Repeat("x", MAX_ECHO_BODY_SIZE+1)
		req := httptest.NewRequest(
			http.MethodPost,
			"/echo",
			strings.NewReader(body),
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf(
				"expected status 413, got %d",
				rec.Code,
			)
		}

		if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
			t.Fatalf("expected problem JSON, got %q", got)
		}

		var problem ProblemResponse
		if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
			t.Fatalf("failed to decode problem response: %v", err)
		}

		if problem.Status != http.StatusRequestEntityTooLarge {
			t.Errorf("expected problem status 413, got %d", problem.Status)
		}
	})
}

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "release version",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.0"},
			},
			want: "v1.2.0",
		},
		{
			name: "development build",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
			},
			want: "",
		},
		{
			name: "empty version",
			info: &debug.BuildInfo{},
			want: "",
		},
		{
			name: "missing build information",
			info: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionFromBuildInfo(tt.info); got != tt.want {
				t.Errorf("versionFromBuildInfo() = %q, want %q", got, tt.want)
			}
		})
	}
}
