package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	MAX_ALLOWED_DELAY  = 30 * time.Second
	MAX_ECHO_BODY_SIZE = 1 << 20 // 1 MiB
)

var Version = "development"

func init() {
	// Preserve an explicit version supplied through -ldflags.
	if Version != "development" {
		return
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if version := versionFromBuildInfo(info); version != "" {
		Version = version
	}
}

func versionFromBuildInfo(info *debug.BuildInfo) string {
	if info == nil {
		return ""
	}

	version := info.Main.Version
	if version == "" || version == "(devel)" {
		return ""
	}

	return version
}

type VersionResponse struct {
	Version string `json:"version"`
}

type EchoResponse struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Query   map[string][]string `json:"query"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type ProblemResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

type StatusResponse struct {
	Code         int      `json:"code"`
	Status       string   `json:"status"`
	Location     string   `json:"location,omitempty"`
	Alternatives []string `json:"alternatives,omitempty"`
}
type DelayResponse struct {
	Duration string `json:"duration"`
	Elapsed  string `json:"elapsed"`
}

func writeJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	contentType string,
	value any,
) {
	body, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	body = append(body, '\n')

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)

	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK\n"))
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, "application/json", VersionResponse{
		Version: Version,
	})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	limitedBody := http.MaxBytesReader(
		w,
		r.Body,
		MAX_ECHO_BODY_SIZE,
	)

	bodyBytes, err := io.ReadAll(limitedBody)
	if err != nil {
		status := http.StatusBadRequest
		detail := "Failed to read the request body."

		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			status = http.StatusRequestEntityTooLarge
			detail = "Request body exceeds the 1 MiB limit for /echo."
		}

		writeJSON(w, r, status, "application/problem+json", ProblemResponse{
			Type:     "about:blank",
			Title:    statusDescription(status),
			Status:   status,
			Detail:   detail,
			Instance: r.URL.Path,
		})
		return
	}

	writeJSON(w, r, http.StatusOK, "application/json", EchoResponse{
		Method:  r.Method,
		URL:     r.URL.String(),
		Query:   r.URL.Query(),
		Headers: r.Header,
		Body:    string(bodyBytes),
	})
}

func statusDescription(code int) string {
	if text := http.StatusText(code); text != "" {
		return text
	}

	return "Unknown Status"
}
func handleStatus(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(r.PathValue("code"))
	if err != nil || code < 200 || code > 599 {
		writeJSON(w, r, http.StatusBadRequest, "application/problem+json", ProblemResponse{
			Type:     "about:blank",
			Title:    statusDescription(http.StatusBadRequest),
			Status:   http.StatusBadRequest,
			Detail:   "Fakeway accepts final status codes from 200 through 599. Informational 1xx responses cannot be returned as standalone final responses.",
			Instance: r.URL.Path,
		})
		return
	}

	if code == 305 || code == 306 {
		writeJSON(w, r, http.StatusBadRequest, "application/problem+json", ProblemResponse{
			Type:     "about:blank",
			Title:    statusDescription(http.StatusBadRequest),
			Status:   http.StatusBadRequest,
			Detail:   "Status 305 is deprecated and status 306 is unused. Fakeway does not simulate these legacy status codes.",
			Instance: r.URL.Path,
		})
		return
	}

	statusText := statusDescription(code)

	if handleSpecialStatus(w, r, code, statusText) {
		return
	}

	if code >= 400 {
		writeJSON(w, r, code, "application/problem+json", ProblemResponse{
			Type:     "about:blank",
			Title:    statusDescription(code),
			Status:   code,
			Detail:   statusDetail(code),
			Instance: r.URL.Path,
		})
		return
	}

	writeJSON(w, r, code, "application/json", StatusResponse{
		Code:   code,
		Status: statusText,
	})
}

func handleDelay(w http.ResponseWriter, r *http.Request) {
	durationStr := r.PathValue("duration")

	d, err := time.ParseDuration(durationStr)
	if err != nil || d < 0 || d > MAX_ALLOWED_DELAY {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid duration. Must be a valid Go time duration (e.g., 500ms, 2s) between 0s and 30s.",
		})
		return
	}

	start := time.Now()
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		writeJSON(w, r, http.StatusOK, "application/json", DelayResponse{
			Duration: d.String(),
			Elapsed:  time.Since(start).String(),
		})

	case <-r.Context().Done():
		return
	}
}
