package lib

import (
	"crypto/sha256"

	"github.com/jxskiss/base62"
)

// To implement correct shortening is non-trivial
// Usually, the long url needs to be associated with a global unique id (number)
// Then this id is used for the base62 encoding
// However, implementing a reliable distributed global unique identifier generator is out of the scope of this project
func ToShortURL(longURL string, length int) string {
	hash := sha256.Sum256([]byte(longURL))
	encoded := base62.EncodeToString(hash[:])
	// Return first 'length' characters
	if len(encoded) >= length {
		return encoded[:length]
	}
	// // Pad if needed
	for len(encoded) < length {
		encoded = "0" + encoded
	}
	return encoded
}
