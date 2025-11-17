package oauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	domain "github.com/kleffio/kleff-auth/internal/core/domain/auth"
)

type StateCodec struct {
	key []byte
}

func NewStateCodec(secretKey string) (*StateCodec, error) {
	var key []byte

	if decoded, err := base64.StdEncoding.DecodeString(secretKey); err == nil {
		key = decoded
	} else {
		key = []byte(secretKey)
	}

	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	return &StateCodec{key: key}, nil
}

func (c *StateCodec) Encode(state *domain.OAuthState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (c *StateCodec) Decode(stateStr string) (*domain.OAuthState, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(stateStr)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("invalid state")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	data, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var state domain.OAuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}
