package httpapi

import "net/http"

func NewHandler() http.Handler {
	return setupRoutes()
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /version", handleVersion)
	mux.HandleFunc("/echo", handleEcho)

	mux.HandleFunc("/status/{code}", handleStatus)

	mux.HandleFunc("GET /delay/{duration}", handleDelay)

	return corsMiddleware(loggingMiddleware(mux))
}
