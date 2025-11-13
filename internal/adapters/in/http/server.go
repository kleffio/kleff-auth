package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/kleff-auth/internal/application/auth"
)

func NewRouter(svc *auth.Service) http.Handler {
	h := &AuthHandlers{SVC: svc}
	mw := &AuthMiddleware{Tokens: svc.Tokens}

	r := chi.NewRouter()

	r.Route("/v1/auth", func(r chi.Router) {
		r.Get("/.well-known/jwks.json", h.JWKS)
		r.Post("/signup", ErrorMiddleware(h.SignUp))
		r.Post("/signin", ErrorMiddleware(h.SignIn))
		r.Post("/refresh", ErrorMiddleware(h.Refresh))
		r.Post("/logout", ErrorMiddleware(h.Logout))
		r.Post("/logout-all", ErrorMiddleware(h.LogoutAll))
		r.Group(func(r chi.Router) {
			r.Use(mw.WithAuth)
			r.Get("/me", ErrorMiddleware(h.Me))
		})
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	return r
}
