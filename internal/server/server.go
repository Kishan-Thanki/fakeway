package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kishan-thanki/fakeway/internal/httpapi"
)

const shutdownTimeout = 30 * time.Second

// Run starts the Fakeway HTTP server and gracefully shuts it down
// when an interrupt or termination signal is received.
func Run() error {
	cfg := parseFlags()

	server := &http.Server{
		Addr:         cfg.addr,
		Handler:      httpapi.NewHandler(),
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
		IdleTimeout:  cfg.idleTimeout,
	}

	sigCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	log.Printf("Starting fakeway server on %s", cfg.addr)

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil

	case <-sigCtx.Done():
		log.Println("Shutdown signal received. Closing fakeway server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("server forced shutdown error: %w", err)
		}

		if err := <-serverErr; err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}

		log.Println("Server stopped cleanly.")
		return nil
	}
}
