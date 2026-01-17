package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

type contextKey string

const requestIDContextKey contextKey = RequestIDKey

// RequestID middleware extracts or generates a request ID and stores it in context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in response header for client correlation
		w.Header().Set(RequestIDHeader, requestID)

		// Store in context for downstream use
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the request (context first, then header).
func GetRequestID(r *http.Request) string {
	// Try context first (set by middleware)
	if requestID, ok := r.Context().Value(requestIDContextKey).(string); ok {
		return requestID
	}
	// Fallback to header
	return r.Header.Get(RequestIDHeader)
}

// GetRequestIDFromContext retrieves the request ID from a context directly.
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
