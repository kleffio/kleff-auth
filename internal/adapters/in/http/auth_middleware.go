package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/kleffio/kleff-auth/internal/core/port/auth"
)

type ctxKey string

const (
	ctxUserID   ctxKey = "user_id"
	ctxTenantID ctxKey = "tenant_id"
)

type AuthMiddleware struct {
	Tokens auth.TokenSignerPort
}

func (m *AuthMiddleware) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := extractBearer(r.Header.Get("Authorization"))
		if tok == "" {
			if c, err := r.Cookie("access_token"); err == nil {
				tok = strings.TrimSpace(c.Value)
			}
		}

		if tok == "" {
			respondWithError(w, http.StatusUnauthorized, "missing token", nil)
			return
		}

		sub, tid, err := m.Tokens.ParseAccess(tok)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid token", nil)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, sub)
		ctx = context.WithValue(ctx, ctxTenantID, tid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearer(h string) string {
	if h == "" {
		return ""
	}
	low := strings.ToLower(h)
	if !strings.HasPrefix(low, "bearer ") {
		return ""
	}
	return strings.TrimSpace(h[7:])
}
