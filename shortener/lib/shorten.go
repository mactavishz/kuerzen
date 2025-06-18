package lib

import (
	"crypto/sha256"

	"github.com/jxskiss/base62"
)

// URL shortening using hash-based approach with collision trade-offs
//
// COLLISION ANALYSIS:
// - 8 chars of base62 = 62^8 ≈ 218 trillion possible values
// - Collision probability follows birthday paradox:
//   - 1K URLs: ~0.000002% collision chance
//   - 100K URLs: ~0.0023% collision chance
//   - 1M URLs: ~0.23% collision chance
//
// PRODUCTION ALTERNATIVES:
// The robust approach uses a reliable distributed counter/snowflake ID:
// 1. Generate globally unique sequential ID for each URL
// 2. Base62-encode the ID (guarantees no collisions)
//
// This hash-based approach is only acceptable for:
// - Development/prototyping environments
// - Applications with roughly 1M URLs where ~0.23% collision risk is tolerable
// - Systems with collision detection and retry logic
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
