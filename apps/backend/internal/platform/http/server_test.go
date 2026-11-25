package http

import (
	"context"
	"io"
	"log/slog"
	nethttp "net/http"
	"testing"
	"time"
)

func TestServer_RunShutsDownOnContextCancel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer("127.0.0.1:0", logger, nethttp.NewServeMux())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not shut down")
	}
}
