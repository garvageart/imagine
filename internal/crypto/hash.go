package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
)

// CreateHash creates a SHA-256 hash of the provided data.
// Returns a 32-byte hash.
func CreateHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyHash compares a hash against the hash of the provided data.
// Returns true if the hashes match.
func VerifyHash(data []byte, hash []byte) bool {
	expected := CreateHash(data)
	return subtle.ConstantTimeCompare(expected, hash) == 1
}
