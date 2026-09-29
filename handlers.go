package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const maxAllowedDelay = 30 * time.Second

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK\n"))
}

type EchoResponse struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Query   map[string][]string `json:"query"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	echoData := EchoResponse{
		Method:  r.Method,
		URL:     r.URL.String(),
		Query:   r.URL.Query(),
		Headers: r.Header,
		Body:    string(bodyBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(echoData)
}

type StatusResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	codeStr := r.PathValue("code")

	code, err := strconv.Atoi(codeStr)
	if err != nil || code < 100 || code > 599 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid status code. Must be an integer between 100 and 599.",
		})
		return
	}

	statusText := http.StatusText(code)
	if statusText == "" {
		statusText = "Unknown Status"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(StatusResponse{
		Code:   code,
		Status: statusText,
	})
}

type DelayResponse struct {
	Duration string `json:"duration"`
	Elapsed  string `json:"elapsed"`
}

func handleDelay(w http.ResponseWriter, r *http.Request) {
	durationStr := r.PathValue("duration")

	d, err := time.ParseDuration(durationStr)
	if err != nil || d < 0 || d > maxAllowedDelay {
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(DelayResponse{
			Duration: d.String(),
			Elapsed:  time.Since(start).String(),
		})
	case <-r.Context().Done():
		return
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (srw *statusResponseWriter) WriteHeader(code int) {
	srw.statusCode = code
	srw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := newStatusResponseWriter(w)

		next.ServeHTTP(srw, r)

		log.Printf("[%s] %s %d %s", r.Method, r.URL.Path, srw.statusCode, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("/echo", handleEcho)
	mux.HandleFunc("GET /status/{code}", handleStatus)
	mux.HandleFunc("GET /delay/{duration}", handleDelay)

	return corsMiddleware(loggingMiddleware(mux))
}
