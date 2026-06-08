// Package hash provides bcrypt for passwords and SHA-256 for opaque tokens.
package hash

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const BCryptCost = 12

func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), BCryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ComparePassword(hashed, plain string) bool {
	if hashed == "" || plain == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

func SHA256(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

func SHA256Hex(s string) string {
	return hex.EncodeToString(SHA256(s))
}
