package http

import (
	"encoding/json"
	"net/http"

	app "github.com/kleffio/kleff-auth/internal/application/auth"
)

type signUpReq struct {
	Tenant   string         `json:"tenant"`
	Email    *string        `json:"email"`
	Username *string        `json:"username"`
	Password string         `json:"password"`
	Attrs    map[string]any `json:"attrs"`
}

type signInReq struct {
	Tenant     string `json:"tenant"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var req signUpReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "bad_json"})
		return
	}

	attrs, _ := json.Marshal(req.Attrs)
	_, tok, err := s.auth.SignUp(r.Context(), app.SignUpInput{
		Tenant: req.Tenant, Email: req.Email, Username: req.Username,
		Password: req.Password, AttrsJSON: attrs,
	})

	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 201, map[string]any{"session": tok})
}

func (s *Server) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var req signInReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "bad_json"})
		return
	}

	_, email, username, tok, err := s.auth.SignIn(r.Context(), app.SignInInput{
		Tenant: req.Tenant, Identifier: req.Identifier, Password: req.Password,
	})

	if err != nil {
		writeJSON(w, 401, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, 200, map[string]any{"user": map[string]any{"email": email, "username": username}, "session": tok})
}
