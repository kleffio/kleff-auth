package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/kleffio/kleff-auth/internal/application/auth"
)

type AuthHandlers struct {
	SVC *auth.Service
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func getClientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return ""
}

func extractRefreshToken(r *http.Request) string {
	if c, err := r.Cookie("refresh_token"); err == nil && c != nil {
		v := strings.TrimSpace(c.Value)
		trimmed := strings.Trim(v, `"'`)
		if trimmed != "" {
			return trimmed
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return ""
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

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

func (h *AuthHandlers) JWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.SVC.JWKS())
}

func (h *AuthHandlers) SignUp(w http.ResponseWriter, r *http.Request) error {
	var in auth.SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return BadRequest("invalid json body")
	}

	in.IP = getClientIP(r)
	in.UserAgent = r.UserAgent()

	userID, tok, err := h.SVC.SignUp(r.Context(), in)
	if err != nil {
		return err
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{
		"user_id": userID,
		"session": tok,
	})
	return nil
}

func (h *AuthHandlers) SignIn(w http.ResponseWriter, r *http.Request) error {
	var in auth.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return BadRequest("invalid json body")
	}

	in.IP = getClientIP(r)
	in.UserAgent = r.UserAgent()

	uid, email, username, tok, err := h.SVC.SignIn(r.Context(), in)
	if err != nil {
		return err
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":       uid,
			"email":    email,
			"username": username,
		},
		"session": tok,
	})
	return nil
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	ip := getClientIP(r)
	ua := r.UserAgent()

	tok, err := h.SVC.RefreshTokens(r.Context(), rt, ua, ip, "")
	if err != nil {
		return err
	}

	setAuthCookies(w, tok.AccessToken, tok.RefreshToken, tok.ExpiresInSec, 60*60*24*30)
	jsonResp(w, http.StatusOK, map[string]any{"session": tok})
	return nil
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	if err := h.SVC.Logout(r.Context(), rt); err != nil {
		return err
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) error {
	rt := extractRefreshToken(r)
	if rt == "" {
		return BadRequest("missing refresh token")
	}

	if err := h.SVC.LogoutAll(r.Context(), rt); err != nil {
		return err
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) error {
	userID, _ := r.Context().Value(ctxUserID).(string)
	tenantID, _ := r.Context().Value(ctxTenantID).(string)

	if userID == "" || tenantID == "" {
		return Unauthorized("missing or invalid authentication")
	}

	email, username, err := h.SVC.Me(r.Context(), tenantID, userID)
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

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
