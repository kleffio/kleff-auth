package crypto

import (
	"crypto/ed25519"
	"log"
)

func main() {
	//Generate both private and public keys using the algorithm
	privKey, pubKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("Error generating Ed25519 keys: %v", err)
	}

	PrivKey := privKey
	PubKey := pubKey
	_ = PrivKey
	_ = PubKey
}
