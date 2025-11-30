package images

import (
	"crypto/sha1"
	"encoding/hex"
)

func CalculateImageChecksum(data []byte) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
