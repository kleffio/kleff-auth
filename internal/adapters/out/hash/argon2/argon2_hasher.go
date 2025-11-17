package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2id struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	SaltLen uint32
	KeyLen  uint32
}

func NewArgon2id() *Argon2id {
	return &Argon2id{
		Time:    3,
		Memory:  64 * 1024,
		Threads: 2,
		SaltLen: 16,
		KeyLen:  32,
	}
}

func (a *Argon2id) Hash(plain string) (string, error) {
	salt := make([]byte, a.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	sum := argon2.IDKey([]byte(plain), salt, a.Time, a.Memory, a.Threads, a.KeyLen)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		a.Memory, a.Time, a.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

func (a *Argon2id) NeedsRehash(encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return true
	}

	params := map[string]string{}
	for _, kv := range strings.Split(parts[3], ",") {
		p := strings.SplitN(strings.TrimSpace(kv), "=", 2)
		if len(p) == 2 {
			params[p[0]] = p[1]
		}
	}

	mv, errM := strconv.ParseUint(params["m"], 10, 32)
	tv, errT := strconv.ParseUint(params["t"], 10, 32)
	pv, errP := strconv.ParseUint(params["p"], 10, 8)
	if errM != nil || errT != nil || errP != nil {
		return true
	}

	if uint32(mv) != a.Memory || uint32(tv) != a.Time || uint8(pv) != a.Threads {
		return true
	}

	decode := func(s string) ([]byte, error) {
		if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
			return b, nil
		}
		return base64.StdEncoding.DecodeString(s)
	}

	salt, err := decode(parts[4])
	if err != nil {
		return true
	}
	sum, err := decode(parts[5])
	if err != nil {
		return true
	}

	maxIntU := uint64(^uint(0) >> 1)
	if uint64(a.SaltLen) > maxIntU || uint64(a.KeyLen) > maxIntU {
		return true
	}
	if len(salt) != int(a.SaltLen) {
		return true
	}
	if len(sum) != int(a.KeyLen) {
		return true
	}

	return false
}

func (a *Argon2id) Verify(plain, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("invalid hash format")
	}

	params := map[string]string{}
	for _, kv := range strings.Split(parts[3], ",") {
		p := strings.SplitN(strings.TrimSpace(kv), "=", 2)
		if len(p) == 2 {
			params[p[0]] = p[1]
		}
	}

	mv, err := strconv.ParseUint(params["m"], 10, 32)
	if err != nil {
		return false, errors.New("invalid m")
	}
	tv, err := strconv.ParseUint(params["t"], 10, 32)
	if err != nil {
		return false, errors.New("invalid t")
	}
	pv, err := strconv.ParseUint(params["p"], 10, 8)
	if err != nil {
		return false, errors.New("invalid p")
	}

	if mv > uint64(^uint32(0)) || tv > uint64(^uint32(0)) {
		return false, errors.New("argon2 params exceed uint32")
	}
	if pv > uint64(^uint8(0)) {
		return false, errors.New("argon2 parallelism exceeds uint8")
	}

	decode := func(s string) ([]byte, error) {
		if b, err := base64.RawStdEncoding.DecodeString(s); err == nil {
			return b, nil
		}
		return base64.StdEncoding.DecodeString(s)
	}

	salt, err := decode(parts[4])
	if err != nil {
		return false, errors.New("invalid salt b64")
	}
	sum, err := decode(parts[5])
	if err != nil {
		return false, errors.New("invalid sum b64")
	}

	maxIntU := uint64(^uint(0) >> 1)
	if uint64(a.KeyLen) > maxIntU {
		return false, errors.New("key length exceeds platform int")
	}
	if len(sum) != int(a.KeyLen) {
		return false, errors.New("unexpected hash length")
	}

	key := argon2.IDKey(
		[]byte(plain),
		salt,
		uint32(tv),
		uint32(mv),
		uint8(pv),
		a.KeyLen,
	)

	return subtle.ConstantTimeCompare(sum, key) == 1, nil
}
