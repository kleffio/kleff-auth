package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/kleff-auth/internal/core/service/auth"
)

type AuthHandlers struct {
	SVC *auth.Service
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func getClientIP(request *http.Request) string {
	if xf := request.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}

func extractRefreshToken(request *http.Request) string {
	if c, err := request.Cookie("refresh_token"); err == nil && c != nil {
		v := strings.TrimSpace(c.Value)
		trimmed := strings.Trim(v, `"'`)
		if trimmed != "" {
			return trimmed
		}
	}

	if request.Method == http.MethodPost || request.Method == http.MethodPut {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return ""
		}

		request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var b refreshBody
		if err := json.Unmarshal(bodyBytes, &b); err == nil {
			v := strings.TrimSpace(b.RefreshToken)
			trimmed := strings.Trim(v, `"'`)
			if trimmed != "" {
				return trimmed
			}
		}
	}

	return ""
}

func setAuthCookies(w http.ResponseWriter, access string, refresh string, accessTTLSeconds int, refreshTTLSeconds int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   accessTTLSeconds,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   refreshTTLSeconds,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/v1/auth",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})
}

func (handler *AuthHandlers) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(handler.SVC.JWKS())
}

func (handler *AuthHandlers) SignUp(w http.ResponseWriter, r *http.Request) error {
	var req signUpRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("invalid json body")
	}

	in := auth.SignUpInput{
		Tenant:    req.Tenant,
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		AttrsJSON: req.AttrsJSON,
		IP:        getClientIP(r),
		UserAgent: r.UserAgent(),
	}

	userID, tok, err := handler.SVC.SignUp(r.Context(), in)
	if err != nil {
		return err
	}

	session := tokenResponseDTO{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresInSec: tok.ExpiresInSec,
		TokenType:    tok.TokenType,
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{
		"user_id": userID,
		"session": session,
	})
	return nil
}

func (handler *AuthHandlers) SignIn(w http.ResponseWriter, r *http.Request) error {
	var req signInRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("invalid json body")
	}

	in := auth.SignInInput{
		Tenant:     req.Tenant,
		Identifier: req.Identifier,
		Password:   req.Password,
		IP:         getClientIP(r),
		UserAgent:  r.UserAgent(),
	}

	uid, email, username, tok, err := handler.SVC.SignIn(r.Context(), in)
	if err != nil {
		return err
	}

	session := tokenResponseDTO{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresInSec: tok.ExpiresInSec,
		TokenType:    tok.TokenType,
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":       uid,
			"email":    email,
			"username": username,
		},
		"session": session,
	})

	return nil
}

func (handler *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	ip := getClientIP(r)
	ua := r.UserAgent()

	tok, err := handler.SVC.RefreshTokens(r.Context(), rt, ua, ip, "")
	if err != nil {
		return err
	}

	session := tokenResponseDTO{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresInSec: tok.ExpiresInSec,
		TokenType:    tok.TokenType,
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{"session": session})

	return nil
}

func (handler *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	if err := handler.SVC.Logout(r.Context(), rt); err != nil {
		return err
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (handler *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	if err := handler.SVC.LogoutAll(r.Context(), rt); err != nil {
		return err
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (handler *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) error {
	userID, _ := r.Context().Value(ctxUserID).(string)
	tenantID, _ := r.Context().Value(ctxTenantID).(string)

	if userID == "" || tenantID == "" {
		return Unauthorized("missing or invalid authentication")
	}

	email, username, err := handler.SVC.Me(r.Context(), tenantID, userID)
	if err != nil {
		return NotFound("user not found")
	}

	jsonResp(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":        userID,
			"tenant_id": tenantID,
			"username":  username,
			"email":     email,
		},
	})
	return nil
}

func (handler *AuthHandlers) OAuthStart(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()

	tenant := q.Get("tenant")
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")

	log.Printf("OAuth Start - provider: %s, tenant: %s, clientID: %s, redirectURI: %s",
		provider, tenant, clientID, redirectURI)

	if tenant == "" || clientID == "" || redirectURI == "" {
		return BadRequest("missing tenant, client_id, or redirect_uri")
	}

	in := auth.OAuthStartInput{
		Provider:    provider,
		Tenant:      tenant,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		IP:          getClientIP(r),
		UserAgent:   r.UserAgent(),
	}

	redirectURL, err := handler.SVC.BuildOAuthRedirectURL(r.Context(), in)
	if err != nil {
		log.Printf("BuildOAuthRedirectURL error: %v", err)
		return err
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

func (handler *AuthHandlers) OAuthCallback(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	stateStr := r.URL.Query().Get("state")

	if code == "" {
		return BadRequest("missing authorization code")
	}
	if stateStr == "" {
		return BadRequest("missing state parameter")
	}

	uid, email, username, tok, err := handler.SVC.HandleOAuthCallback(
		r.Context(), provider, code, stateStr, getClientIP(r), r.UserAgent(),
	)
	if err != nil {
		return err
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)

	st, err := handler.SVC.OAuthState.Decode(stateStr)
	if err != nil {
		log.Printf("OAuthCallback: failed to decode state for redirect: %v", err)

		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			UserID   string           `json:"user_id"`
			Email    *string          `json:"email,omitempty"`
			Username *string          `json:"username,omitempty"`
			Token    tokenResponseDTO `json:"token"`
		}{
			UserID:   uid,
			Email:    email,
			Username: username,
			Token: tokenResponseDTO{
				AccessToken:  tok.AccessToken,
				RefreshToken: tok.RefreshToken,
				ExpiresInSec: tok.ExpiresInSec,
				TokenType:    tok.TokenType,
			},
		}
		return json.NewEncoder(w).Encode(resp)
	}

	if st.RedirectURI == "" {
		st.RedirectURI = "http://localhost:5173/"
	}

	http.Redirect(w, r, st.RedirectURI, http.StatusFound)
	return nil
}

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
