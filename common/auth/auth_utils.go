package auth

import "crypto/rand"

const (
	APIKeyPrefix = "img"
***REMOVED***

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
		"Tempest"***REMOVED***
***REMOVED***

func GenerateRandomBytes(n int***REMOVED*** ([]byte***REMOVED*** {
	b := make([]byte, n***REMOVED***
	_, _ = rand.Read(b***REMOVED***

	return b
***REMOVED***
