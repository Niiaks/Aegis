package router

import (
	"github.com/Niiaks/Aegis/internal/middleware"
	"github.com/Niiaks/Aegis/internal/server"
	"github.com/Niiaks/Aegis/internal/user"
	"github.com/Niiaks/Aegis/internal/wallet"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	User   *user.UserHandler
	Wallet *wallet.WalletHandler
}

func NewRouter(s *server.Server, h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	mw := middleware.NewMiddlewares(s)

	// Apply middleware in order
	r.Use(middleware.RequestID)
	r.Use(mw.Tracing.NewRelicMiddleware())
	r.Use(mw.Tracing.EnhanceTracing)
	r.Use(mw.ContextEnhancer.EnhanceContext)
	r.Use(mw.Global.RequestLogger)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Post("/register", h.User.CreateUser)
		})

		// Wallet routes
		r.Route("/wallets", func(r chi.Router) {
			r.Post("/create", h.Wallet.CreateWallet)
		})

	})

	return r
}
