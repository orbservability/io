package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
)

func ServeHTTP(errChan chan<- error, handlers map[string]http.Handler) *http.Server {
	mux := http.NewServeMux()

	// Range over the map and assign handlers to paths
	for path, handler := range handlers {
		mux.Handle(path, handler)
	}

	httpServer := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	return httpServer
}

func ShutdownHTTP(httpServer *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout())
	defer cancel()

	done := make(chan struct{})
	go func() {
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("HTTP server shutdown error")
		}
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("HTTP server shut down gracefully")
	case <-ctx.Done():
		log.Warn().Msg("HTTP server shutdown timed out")
	}
}
