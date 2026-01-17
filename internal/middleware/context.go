package middleware

import (
	"context"
	"net/http"

	"github.com/Niiaks/Aegis/internal/logger"
	"github.com/Niiaks/Aegis/internal/server"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

type ContextEnhancer struct {
	Server *server.Server
}

func NewContextEnhancer(srv *server.Server) *ContextEnhancer {
	return &ContextEnhancer{
		Server: srv,
	}
}

func (ce *ContextEnhancer) EnhanceContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r)

		//enhance logger with context
		contextLogger := ce.Server.Logger.With().
			Str("request_id", requestID).
			Str("ip", r.RemoteAddr).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Logger()

		if txn := newrelic.FromContext(r.Context()); txn != nil {
			contextLogger = logger.WithTraceContext(contextLogger, txn)
		}

		if userID := ce.extractUserId(r); userID != "" {
			contextLogger = contextLogger.With().Str("user_id", userID).Logger()
		}

		//set enhanced logger in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, LoggerKey, &contextLogger)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})

}

func (ce *ContextEnhancer) extractUserId(r *http.Request) string {
	if userID, ok := r.Context().Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// func (ce *ContextEnhancer) ExtractUserRole(r *http.Request) string {
// 	if userRole, ok := r.Context().Value(UserRoleKey).(string); ok {
// 		return userRole
// 	}
// 	return ""
// }

// GetLogger retrieves the logger from the context.
func GetLogger(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger); ok {
		return logger
	}
	logger := zerolog.Nop()
	return &logger
}
