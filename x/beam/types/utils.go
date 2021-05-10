package types

import (
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
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
func GenerateHashFromString(secret string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// CompareHashAndString is used to verify that a string matches a provided hash
func CompareHashAndString(hash string, str string) bool {
	res := bcrypt.CompareHashAndPassword([]byte(hash), []byte(str))
	return res == nil
}