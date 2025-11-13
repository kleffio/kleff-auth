package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kleffio/kleff-auth/internal/core/service/auth"
)

func NewRouter(svc *auth.Service) http.Handler {
	h := &AuthHandlers{SVC: svc}
	mw := &AuthMiddleware{Tokens: svc.Tokens}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://kleff.io", "https://kleff.ca", "https://kleff.app", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Route("/v1/auth", func(r chi.Router) {
		r.Get("/.well-known/jwks.json", h.JWKS)
		r.Post("/signup", ErrorMiddleware(h.SignUp))
		r.Post("/signin", ErrorMiddleware(h.SignIn))
		r.Post("/refresh", ErrorMiddleware(h.Refresh))
		r.Post("/logout", ErrorMiddleware(h.Logout))
		r.Post("/logout-all", ErrorMiddleware(h.LogoutAll))

		r.Get("/oauth/{provider}/start", ErrorMiddleware(h.OAuthStart))
		r.Get("/oauth/{provider}/callback", ErrorMiddleware(h.OAuthCallback))

		r.Group(func(r chi.Router) {
			r.Use(mw.WithAuth)
			r.Get("/me", ErrorMiddleware(h.Me))
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	return r
}
