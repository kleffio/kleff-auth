package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
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

	return &EddsaSigner{KID: newKID(), Priv: priv, Pub: pub, Issuer: issuer}, nil
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

type claims struct {
	Sub   string `json:"sub"`
	Tid   string `json:"tid"`
	Scope string `json:"scope"`
	JTI   string `json:"jti"`
	jwt.RegisteredClaims
}

func (s *EddsaSigner) IssueAccess(sub, tid, scope, jti string, ttl time.Duration) (string, error) {
	c := claims{
		Sub: sub, Tid: tid, Scope: scope, JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	t.Header["kid"] = s.KID

	return t.SignedString(s.Priv)
}

func (s *EddsaSigner) NewRefresh() (string, string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	raw := base64.RawURLEncoding.EncodeToString(b)

	return raw, raw, nil
}
