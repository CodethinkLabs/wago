package util

import "golang.org/x/crypto/ed25519"

// converts a public key to a fixed-size byte array
func ToBytes(key ed25519.PublicKey) [ed25519.PublicKeySize]byte {
	var bytes [ed25519.PublicKeySize]byte
	copy(bytes[:], key)
	return bytes
}
