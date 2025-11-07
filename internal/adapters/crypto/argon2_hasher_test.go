package crypto

import "testing"

func TestArgon2id(t *testing.T) {
	a := NewArgon2id()
	hash, _ := a.Hash("test123")

	ok, err := a.Verify("test123", hash)
	if err != nil || !ok {
		t.Fatalf("verify failed: %v", err)
	}

	t.Log("hash:", hash)
}
