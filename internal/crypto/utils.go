package crypto

import (
	"crypto/rand"
	"errors"
)

// MustGenerateRandomBytes generates n random bytes using crypto/rand.
// It returns an error if the system's secure random
// number generator fails or n is less than or equal to zero.
func GenerateRandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("number of bytes must be greater than zero")
	}

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// MustGenerateRandomBytes generates n random bytes using crypto/rand.
// It does not return an error; if the system's secure random
// number generator fails, it will panic
func MustGenerateRandomBytes(n int) []byte {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		panic(err)
	}

	return b
}