package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	HMACSecret = func() []byte {
		secret := os.Getenv("HMAC_SIGNITURE")
		secretBytes, _ := hex.DecodeString(secret)
		return secretBytes
	}()
)

func CreateHash(data []byte) []byte {
	hmac := hmac.New(sha256.New, HMACSecret)
	hmac.Write(data)
	dataHmac := hmac.Sum(nil)

	return dataHmac
}

func VerifyHash(data []byte, signitureData []byte) bool {
	dataHmac := CreateHash(data)
	return hmac.Equal(dataHmac, signitureData)
}
