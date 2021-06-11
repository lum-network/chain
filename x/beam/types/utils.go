package types

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

// GenerateSecureToken Generate a secure random token of the given length and return string
func GenerateSecureToken(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// GenerateHashFromString is used to generate a hashed version of the argument passed string
func GenerateHashFromString(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// CompareHashAndString is used to verify that a string matches a provided hash
func CompareHashAndString(hash string, secret string) bool {
	hashedStr := GenerateHashFromString(secret)
	return hex.EncodeToString(hashedStr) == hash
}