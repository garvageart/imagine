package uid

import (
	"regexp"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	Base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DefaultLength  = 24 // note: might change
)

var (
	AlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`) // TODO: benchmark this against enumerating
)

// Generates a new UID at the default Imagine length
// with Base62 characters (0-9a-zA-Z)
func Generate() (string, error) {
	return gonanoid.Generate(Base62Alphabet, DefaultLength)
}

func IsValid(str string) bool {
	if len(str) == DefaultLength {
		return true
	}

	if AlphanumericRegex.MatchString(str) {
		return true
	}

	return false
}
