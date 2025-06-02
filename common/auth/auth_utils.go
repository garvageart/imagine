package auth

import "crypto/rand" // aliased to avoid conflict with math/rand/v2 if used in the same package elsewhere

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

var(
	APIKeyPrefix = "img"
)

func GenerateRandomBytes(n int) ([]byte) {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return b
}