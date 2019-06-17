package util

import "golang.org/x/crypto/ed25519"

func ToBytes(key ed25519.PublicKey) [32]byte {
	var bytes [ed25519.PublicKeySize]byte
	copy(bytes[:], key)
	return bytes
}
