package middleware

import (
	"net/http"
	"time"

	"github.com/Niiaks/Aegis/internal/server"
)

type Global struct {
	s *server.Server
}

func NewGlobal(s *server.Server) *Global {
	return &Global{
		s: s,
	}
}

// RequestLogger logs incoming requests and their duration.
func (g *Global) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log after request completes
		duration := time.Since(start)
		requestID := GetRequestID(r)

		g.s.Logger.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Msg("request completed")
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
