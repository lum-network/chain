package types

import (
	"encoding/hex"
	"math/rand"
)

// GenerateSecureToken Generate a secure random token of the given length and return string
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
