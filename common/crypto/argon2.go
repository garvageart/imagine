package crypto

import (
	"bytes"
	"imagine/common/auth"

	"golang.org/x/crypto/argon2"
)

type Argon2Hash struct {
	// time represents the number of
	// passed over the specified memory.
	time uint32
	// cpu memory to be used.
	memory uint32
	// threads for parallelism aspect
	// of the algorithm.
	threads uint8
	// keyLen of the generate hash key.
	keyLen uint32
	// saltLen the length of the salt used.
	saltLen int
}

func (a Argon2Hash) GenerateSalt() []byte {
	return auth.GenerateRandomBytes(a.saltLen)
}

func (a Argon2Hash) Hash(password, salt []byte) []byte {
	data := append(salt, []byte(":")...)
	if len(salt) == 0 {
		salt = a.GenerateSalt()
	}

	hash := argon2.IDKey(password, salt, a.time, a.memory, a.threads, a.keyLen)
	return append(data, hash...)
}

func (a Argon2Hash) Verify(hashedValue []byte, password string) bool {
	salt := hashedValue[a.saltLen+len(":") : a.saltLen+len(":")+int(a.keyLen)]
	hash := a.Hash([]byte(password), salt)

	return bytes.Equal(hashedValue[a.saltLen+len(":")+int(a.keyLen):], hash)
}

func CreateArgon2Hash(time uint32, saltLen int, memory uint32, threads uint8, keyLen uint32) *Argon2Hash {
	return &Argon2Hash{
		time:    time,
		saltLen: saltLen,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
	}
}
