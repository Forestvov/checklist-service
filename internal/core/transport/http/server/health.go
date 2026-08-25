package core_http_server

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

func (s *HTTPServer) RegisterHealth(checker HealthChecker) {
	s.mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		s.writeHealthResponse(w, http.StatusOK, "ok")
	})

	s.mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), s.config.HealthCheckTimeout)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			s.log.Warn("readiness check failed", zap.Error(err))
			s.writeHealthResponse(w, http.StatusServiceUnavailable, "unavailable")
			return
		}

		s.writeHealthResponse(w, http.StatusOK, "ok")
	})
}

func (s *HTTPServer) writeHealthResponse(
	w http.ResponseWriter,
	statusCode int,
	status string,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if _, err := fmt.Fprintf(w, "{\"status\":%q}\n", status); err != nil {
		s.log.Error("write health response", zap.Error(err))
	}
}
