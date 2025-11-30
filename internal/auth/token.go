package auth

import (
	"encoding/hex"
	"imagine/internal/crypto"
)

func GenerateAuthToken() string {
	tokenBytes := crypto.MustGenerateRandomBytes(48)
	return hex.EncodeToString(tokenBytes)
}
