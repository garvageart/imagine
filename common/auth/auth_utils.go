package auth

import "crypto/rand"

var (
	// Choose a random horse name because why not
	HorseNames = [...]string{
		"RainstormPockets",
		"Azur",
		"Lindsay",
		"Dawson",
		"Tricks",
		"Eclipse",
		"Chesterfield",
		"Flint",
		"Karma",
		"Wind Chaser",
		"Hally",
		"Bellator",
		"Paladen",
		"Cash",
		"Hazel",
		"Chip",
		"Summoner",
		"Lightning",
		"Tempest",
	}
)

var (
	APIKeyPrefix = "imag"
)

func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return b
}