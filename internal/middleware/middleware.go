package middleware

import (
	"github.com/Niiaks/Aegis/internal/server"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Middlewares struct {
	Global          *Global
	ContextEnhancer *ContextEnhancer
	Tracing         *TracingMiddleware
}

func NewMiddlewares(s *server.Server) *Middlewares {

	var nrApp *newrelic.Application

	if s.LoggerService != nil {
		nrApp = s.LoggerService.GetApplication()
	}

	return &Middlewares{
		Global:          NewGlobal(s),
		ContextEnhancer: NewContextEnhancer(s),
		Tracing:         NewTracing(nrApp),
	}
}
