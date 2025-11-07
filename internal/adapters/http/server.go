package http

import (
	"net/http"

	app "github.com/kleffio/kleff-auth/internal/application/auth"
)

type Server struct {
	auth *app.Service
}

func NewServer(authService *app.Service) *http.ServeMux {
	s := &Server{auth: authService}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /.well-known/jwks.json", s.handleJWKS)

	mux.HandleFunc("POST /v1/auth/signup", s.handleSignUp)
	mux.HandleFunc("POST /v1/auth/signin", s.handleSignIn)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(s.auth.JWKS())
}
