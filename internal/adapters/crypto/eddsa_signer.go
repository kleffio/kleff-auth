// internal/adapters/crypto/eddsa_signer.go
package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type EddsaSigner struct {
	KID    string
	Priv   ed25519.PrivateKey
	Pub    ed25519.PublicKey
	Issuer string
}

func NewInMemorySigner(issuer string) (*EddsaSigner, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &EddsaSigner{
		KID:    newKID(),
		Priv:   priv,
		Pub:    pub,
		Issuer: issuer,
	}, nil
}

func newKID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *EddsaSigner) JWKS() []byte {
	type jwk struct{ Kty, Crv, X, Kid, Alg, Use string }
	type jwks struct {
		Keys []jwk `json:"keys"`
	}

	out := jwks{Keys: []jwk{{
		Kty: "OKP", Crv: "Ed25519",
		X:   base64.RawURLEncoding.EncodeToString(s.Pub),
		Kid: s.KID, Alg: "EdDSA", Use: "sig",
	}}}
	b, _ := json.Marshal(out)
	return b
}

type accessClaims struct {
	Sub   string `json:"sub"`
	Tid   string `json:"tid"`
	Scope string `json:"scope"`
	JTI   string `json:"jti"`
	jwt.RegisteredClaims
}

func (s *EddsaSigner) IssueAccess(sub, tid, scope, jti string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := accessClaims{
		Sub: sub, Tid: tid, Scope: scope, JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			Audience:  []string{"kleff-auth"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	t.Header["kid"] = s.KID
	return t.SignedString(s.Priv)
}

func (s *EddsaSigner) ParseAccess(token string) (string, string, error) {
	if token == "" {
		return "", "", errors.New("missing token")
	}
	t, err := jwt.ParseWithClaims(
		token,
		&accessClaims{},
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodEdDSA {
				return nil, errors.New("unexpected alg")
			}
			return s.Pub, nil
		},
		jwt.WithIssuer(s.Issuer),
		jwt.WithAudience("kleff-auth"),
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	)
	if err != nil || !t.Valid {
		return "", "", errors.New("invalid token")
	}
	claims, ok := t.Claims.(*accessClaims)
	if !ok || claims.Sub == "" || claims.Tid == "" {
		return "", "", errors.New("bad claims")
	}
	return claims.Sub, claims.Tid, nil
}
