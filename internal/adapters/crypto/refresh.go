package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type RefreshCodec struct {
	Hasher *Argon2id
}

func (c *RefreshCodec) Generate() ([]byte, []byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, nil, err
	}

	enc, err := c.Hasher.Hash(base64.RawURLEncoding.EncodeToString(secret))
	if err != nil {
		return nil, nil, err
	}

	return secret, []byte(enc), nil
}

func (c *RefreshCodec) Verify(hash, raw []byte) error {
	ok, err := c.Hasher.Verify(base64.RawURLEncoding.EncodeToString(raw), string(hash))
	if err != nil || !ok {
		return errors.New("invalid refresh")
	}

	return nil
}

func (c *RefreshCodec) Encode(sessionID uuid.UUID, secret []byte) string {
	return base64.RawURLEncoding.EncodeToString(sessionID[:]) + "." +
		base64.RawURLEncoding.EncodeToString(secret)
}

func (c *RefreshCodec) Parse(rt string) (uuid.UUID, []byte, error) {
	parts := strings.Split(rt, ".")
	if len(parts) != 2 {
		return uuid.Nil, nil, fmt.Errorf("bad refresh token format: expected 2 parts, got %d", len(parts))
	}

	idPart := parts[0]
	secPart := parts[1]

	idb, err := base64.RawURLEncoding.DecodeString(idPart)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("bad session id encoding: %w", err)
	}

	if len(idb) != 16 {
		return uuid.Nil, nil, fmt.Errorf("bad session id length: expected 16, got %d", len(idb))
	}

	secret, err := base64.RawURLEncoding.DecodeString(secPart)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("bad secret encoding: %w", err)
	}

	sid, err := uuid.FromBytes(idb)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("invalid uuid bytes: %w", err)
	}

	return sid, secret, nil
}
