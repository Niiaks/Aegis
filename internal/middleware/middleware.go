package middleware

import (
	"github.com/Niiaks/Aegis/internal/server"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Middleware struct {
	Global          *Global
	ContextEnhancer *ContextEnhancer
	Tracing         *Tracing
}

func NewMiddleware(s *server.Server) *Middleware {

	var nrApp *newrelic.Application

	if s.LoggerService != nil {
		nrApp = s.LoggerService.GetApplication()
	}

	return &Middleware{
		Global:          NewGlobal(s),
		ContextEnhancer: NewContextEnhancer(s),
		Tracing:         NewTracing(nrApp),
	}
}
