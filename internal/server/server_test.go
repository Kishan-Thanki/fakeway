package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunServerReturnsStartupError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve address for startup error test: %v", err)
	}
	defer listener.Close()

	cfg := testConfig(listener.Addr().String())
	err = runServer(cfg, context.Background(), http.NotFoundHandler())
	if err == nil {
		t.Fatal("expected server startup to fail when the address is already in use")
	}

	if !strings.Contains(err.Error(), "HTTP server error") {
		t.Errorf("expected wrapped HTTP server error, got %v", err)
	}

	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		t.Errorf("expected underlying network operation error, got %T: %v", err, err)
	}
}

func TestRunServerGracefulShutdownWaitsForInflightRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve ephemeral address: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("release ephemeral address: %v", err)
	}

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() { close(releaseRequest) })
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slow" {
			http.NotFound(w, r)
			return
		}

		close(requestStarted)
		<-releaseRequest
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "completed")
	})

	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		defer close(serverDone)
		serverDone <- runServer(testConfig(addr), ctx, handler)
	}()

	requestResult := make(chan struct {
		status int
		body   string
		err    error
	}, 1)
	go func() {
		client := &http.Client{Timeout: 4 * time.Second}
		resp, err := client.Get("http://" + addr + "/slow")
		if err != nil {
			requestResult <- struct {
				status int
				body   string
				err    error
			}{err: err}
			return
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		requestResult <- struct {
			status int
			body   string
			err    error
		}{status: resp.StatusCode, body: string(body), err: readErr}
	}()

	defer func() {
		cancel()
		release()
		select {
		case <-serverDone:
		case <-time.After(5 * time.Second):
			t.Error("server goroutine did not exit during cleanup")
		}
	}()

	select {
	case <-requestStarted:
	case err, ok := <-serverDone:
		if ok {
			t.Fatalf("server exited before request started: %v", err)
		}
		t.Fatal("server exited before request started")
	case <-time.After(3 * time.Second):
		t.Fatal("request did not reach the server in time")
	}

	cancel()
	select {
	case err, ok := <-serverDone:
		if ok {
			t.Fatalf("server stopped before the in-flight request completed: %v", err)
		}
		t.Fatal("server stopped before the in-flight request completed")
	default:
		// Shutdown should wait while the handler is deliberately blocked.
	}

	release()

	select {
	case result := <-requestResult:
		if result.err != nil {
			t.Fatalf("in-flight request failed during shutdown: %v", result.err)
		}
		if result.status != http.StatusOK {
			t.Errorf("expected in-flight request status 200, got %d", result.status)
		}
		if result.body != "completed" {
			t.Errorf("expected body %q, got %q", "completed", result.body)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("in-flight request did not complete after it was released")
	}

	select {
	case err, ok := <-serverDone:
		if !ok {
			return
		}
		if err != nil {
			t.Fatalf("expected graceful shutdown to return nil, got %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("server did not finish graceful shutdown")
	}
}

func testConfig(addr string) config {
	return config{
		addr:         addr,
		readTimeout:  2 * time.Second,
		writeTimeout: 2 * time.Second,
		idleTimeout:  2 * time.Second,
	}
}
