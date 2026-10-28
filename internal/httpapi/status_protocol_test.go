package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusProtocolSemantics(t *testing.T) {
	tests := []struct {
		name                string
		method              string
		path                string
		requestHeaders      map[string]string
		expectedCode        int
		expectedHeaders     map[string]string
		expectedContentType string
		expectedBody        string
		noBody              bool
		noContentType       bool
	}{
		{
			name:         "Created",
			method:       http.MethodPost,
			path:         "/status/201?location=%2Fresources%2F123",
			expectedCode: http.StatusCreated,
			expectedHeaders: map[string]string{
				"Location": "/resources/123",
			},
			expectedContentType: "application/json",
		},
		{
			name:          "No Content",
			method:        http.MethodGet,
			path:          "/status/204",
			expectedCode:  http.StatusNoContent,
			noBody:        true,
			noContentType: true,
		},
		{
			name:         "Reset Content",
			method:       http.MethodGet,
			path:         "/status/205",
			expectedCode: http.StatusResetContent,
			expectedHeaders: map[string]string{
				"Content-Length": "0",
			},
			noBody:        true,
			noContentType: true,
		},
		{
			name:   "Partial Content",
			method: http.MethodGet,
			path:   "/status/206",
			requestHeaders: map[string]string{
				"Range": "bytes=0-5",
			},
			expectedCode: http.StatusPartialContent,
			expectedHeaders: map[string]string{
				"Content-Range": "bytes 0-5/22",
				"Accept-Ranges": "bytes",
			},
			expectedContentType: "text/plain",
			expectedBody:        "Fakewa",
		},
		{
			name:         "Multiple Choices",
			method:       http.MethodGet,
			path:         "/status/300",
			expectedCode: http.StatusMultipleChoices,
			expectedHeaders: map[string]string{
				"Location": "/health",
			},
			expectedContentType: "application/json",
		},
		{
			name:         "Moved Permanently",
			method:       http.MethodGet,
			path:         "/status/301?location=%2Fversion",
			expectedCode: http.StatusMovedPermanently,
			expectedHeaders: map[string]string{
				"Location": "/version",
			},
		},
		{
			name:         "Found",
			method:       http.MethodGet,
			path:         "/status/302?location=%2Fversion",
			expectedCode: http.StatusFound,
			expectedHeaders: map[string]string{
				"Location": "/version",
			},
		},
		{
			name:         "See Other",
			method:       http.MethodPost,
			path:         "/status/303?location=%2Fversion",
			expectedCode: http.StatusSeeOther,
			expectedHeaders: map[string]string{
				"Location": "/version",
			},
		},
		{
			name:         "Temporary Redirect",
			method:       http.MethodPost,
			path:         "/status/307?location=%2Fversion",
			expectedCode: http.StatusTemporaryRedirect,
			expectedHeaders: map[string]string{
				"Location": "/version",
			},
		},
		{
			name:         "Permanent Redirect",
			method:       http.MethodPost,
			path:         "/status/308?location=%2Fversion",
			expectedCode: http.StatusPermanentRedirect,
			expectedHeaders: map[string]string{
				"Location": "/version",
			},
		},
		{
			name:   "Not Modified",
			method: http.MethodGet,
			path:   "/status/304",
			requestHeaders: map[string]string{
				"If-None-Match": fakeStatusETag(),
			},
			expectedCode: http.StatusNotModified,
			expectedHeaders: map[string]string{
				"ETag":          fakeStatusETag(),
				"Cache-Control": "no-cache",
			},
			noBody:        true,
			noContentType: true,
		},
		{
			name:         "Unauthorized",
			method:       http.MethodGet,
			path:         "/status/401",
			expectedCode: http.StatusUnauthorized,
			expectedHeaders: map[string]string{
				"WWW-Authenticate": `Basic realm="Fakeway"`,
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodPost,
			path:         "/status/405",
			expectedCode: http.StatusMethodNotAllowed,
			expectedHeaders: map[string]string{
				"Allow": "GET, HEAD, PUT, PATCH, DELETE, OPTIONS",
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Request Timeout",
			method:       http.MethodGet,
			path:         "/status/408",
			expectedCode: http.StatusRequestTimeout,
			expectedHeaders: map[string]string{
				"Connection": "close",
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Proxy Authentication Required",
			method:       http.MethodGet,
			path:         "/status/407",
			expectedCode: http.StatusProxyAuthRequired,
			expectedHeaders: map[string]string{
				"Proxy-Authenticate": `Basic realm="Fakeway"`,
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:   "Range Not Satisfiable",
			method: http.MethodGet,
			path:   "/status/416",
			requestHeaders: map[string]string{
				"Range": "bytes=999-",
			},
			expectedCode: http.StatusRequestedRangeNotSatisfiable,
			expectedHeaders: map[string]string{
				"Content-Range": "bytes */22",
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Upgrade Required",
			method:       http.MethodGet,
			path:         "/status/426",
			expectedCode: http.StatusUpgradeRequired,
			expectedHeaders: map[string]string{
				"Connection": "Upgrade",
				"Upgrade":    "websocket",
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Too Many Requests",
			method:       http.MethodGet,
			path:         "/status/429",
			expectedCode: http.StatusTooManyRequests,
			expectedHeaders: map[string]string{
				"Retry-After": "60",
			},
			expectedContentType: "application/problem+json",
		},
		{
			name:         "Service Unavailable",
			method:       http.MethodGet,
			path:         "/status/503",
			expectedCode: http.StatusServiceUnavailable,
			expectedHeaders: map[string]string{
				"Retry-After": "60",
			},
			expectedContentType: "application/problem+json",
		},
	}

	router := setupRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			for name, value := range tt.requestHeaders {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Fatalf("expected status %d, got %d", tt.expectedCode, rec.Code)
			}

			for name, value := range tt.expectedHeaders {
				if got := rec.Header().Get(name); got != value {
					t.Errorf("header %s = %q, want %q", name, got, value)
				}
			}

			if tt.expectedContentType != "" {
				got := rec.Header().Get("Content-Type")

				if strings.HasPrefix(tt.expectedContentType, "text/plain") {
					if !strings.HasPrefix(got, tt.expectedContentType) {
						t.Errorf("Content-Type = %q, want prefix %q", got, tt.expectedContentType)
					}
				} else if got != tt.expectedContentType {
					t.Errorf("Content-Type = %q, want %q", got, tt.expectedContentType)
				}
			}

			if tt.noContentType && rec.Header().Get("Content-Type") != "" {
				t.Errorf("expected no Content-Type, got %q", rec.Header().Get("Content-Type"))
			}

			if tt.noBody && rec.Body.Len() != 0 {
				t.Errorf("expected no response body, got %q", rec.Body.String())
			}

			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestStatusProtocolConditions(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
	}{
		{
			name:   "Created rejects GET",
			method: http.MethodGet,
			path:   "/status/201",
		},
		{
			name:   "Partial Content requires Range",
			method: http.MethodGet,
			path:   "/status/206",
		},
		{
			name:   "Partial Content rejects POST",
			method: http.MethodPost,
			path:   "/status/206",
			headers: map[string]string{
				"Range": "bytes=0-5",
			},
		},
		{
			name:   "Not Modified requires a matching validator",
			method: http.MethodGet,
			path:   "/status/304",
		},
		{
			name:   "Not Modified rejects an unmatched validator",
			method: http.MethodGet,
			path:   "/status/304",
			headers: map[string]string{
				"If-None-Match": `"another-etag"`,
			},
		},
		{
			name:   "Range Not Satisfiable requires Range",
			method: http.MethodGet,
			path:   "/status/416",
		},
		{
			name:   "Range Not Satisfiable rejects a satisfiable range",
			method: http.MethodGet,
			path:   "/status/416",
			headers: map[string]string{
				"Range": "bytes=0-5",
			},
		},
		{
			name:   "Redirect rejects external location",
			method: http.MethodGet,
			path:   "/status/302?location=https%3A%2F%2Fexample.com",
		},
		{
			name:   "Created rejects external location",
			method: http.MethodPost,
			path:   "/status/201?location=https%3A%2F%2Fexample.com",
		},
	}

	router := setupRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", rec.Code)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("expected problem content type, got %q", got)
			}

			var resp ProblemResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode problem response: %v", err)
			}

			if resp.Status != http.StatusBadRequest {
				t.Errorf("problem status = %d, want 400", resp.Status)
			}
		})
	}
}
