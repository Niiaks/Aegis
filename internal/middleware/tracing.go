package middleware

import (
	"net/http"

	"github.com/newrelic/go-agent/v3/integrations/nrgochi"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Tracing struct {
	nrApp *newrelic.Application
}

func NewTracing(nrApp *newrelic.Application) *Tracing {
	return &Tracing{
		nrApp: nrApp,
	}
}

func (t *Tracing) NewRelicMiddleware() func(http.Handler) http.Handler {
	if t.nrApp == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return nrgochi.Middleware(t.nrApp)
}

// EnhanceTracing adds custom attributes to the New Relic transaction.
func (t *Tracing) EnhanceTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := newrelic.FromContext(r.Context())
		if txn == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Add request attributes
		txn.AddAttribute("http.real_ip", r.RemoteAddr)
		txn.AddAttribute("http.user_agent", r.UserAgent())

		if requestID := GetRequestID(r); requestID != "" {
			txn.AddAttribute("request.id", requestID)
		}

		if userID, ok := r.Context().Value(UserIDKey).(string); ok && userID != "" {
			txn.AddAttribute("user.id", userID)
		}

		next.ServeHTTP(w, r)
	})
}
