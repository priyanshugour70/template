// Package tokens generates cryptographically-random opaque tokens.
package tokens

import (
	"crypto/rand"
	"encoding/base64"
)

// New returns a URL-safe base64 opaque token of byteLen random bytes.
// Default 32 bytes (~43 base64 chars) — sufficient entropy for refresh tokens
// and invite/reset tokens.
func New(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 32
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
