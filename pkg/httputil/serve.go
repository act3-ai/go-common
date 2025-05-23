package httputil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/act3-ai/go-common/pkg/logger"
)

// TODO support HTTPS (TLS) with srv.ListenAndServeTLS(certFile, keyFile)

// Serve will run the http server until the context is done.  Then it gracefully shutdown.
func Serve(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	log := logger.FromContext(ctx)

	// Run our server in a goroutine so that it doesn't block.
	// TODO use structured concurrency here (conc.WaitGroup)
	go func() {
		log.InfoContext(ctx, "Listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.ErrorContext(ctx, "serve failed", "error", err)
				panic(err)
			}
		}
	}()

	// graceful shutdown adapted from https://github.com/gorilla/mux#graceful-shutdown (and Telemetry)

	<-ctx.Done()
	log.InfoContext(ctx, "Graceful HTTP server shutdown requested")

	// Create a deadline to wait for.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.InfoContext(ctx, "Waiting for graceful shutdown", slog.Duration("timeout", timeout))
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(timeoutCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return nil
}
