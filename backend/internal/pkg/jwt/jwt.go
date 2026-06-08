// Package jwt issues and verifies HS256 access tokens for the auth module.
// Refresh tokens are opaque (see internal/pkg/tokens) and stored hashed in DB.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalid = errors.New("invalid token")
	ErrExpired = errors.New("token expired")
)

// Claims is the payload of the access token.
type Claims struct {
	UserID         uuid.UUID `json:"sub_uid"`
	TenantID       uuid.UUID `json:"tid"`
	OrganizationID uuid.UUID `json:"oid,omitempty"`
	MembershipID   uuid.UUID `json:"mid,omitempty"`
	Email          string    `json:"email,omitempty"`
	IsSuperAdmin   bool      `json:"sa,omitempty"`
	jwt.RegisteredClaims
}

type Signer struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewSigner(secret string, ttl time.Duration, issuer string) *Signer {
	if issuer == "" {
		issuer = "app"
	}
	return &Signer{secret: []byte(secret), ttl: ttl, issuer: issuer}
}

func (s *Signer) TTL() time.Duration { return s.ttl }

// Issue mints a signed access token. Returns the token string, expiry time, and jti.
func (s *Signer) Issue(c Claims) (token string, expiresAt time.Time, jti string, err error) {
	now := time.Now()
	exp := now.Add(s.ttl)
	jti = uuid.NewString()
	c.RegisteredClaims = jwt.RegisteredClaims{
		ID:        jti,
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
		Subject:   c.UserID.String(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	signed, err := t.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("sign token: %w", err)
	}
	return signed, exp, jti, nil
}

func (s *Signer) Verify(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, ErrInvalid
	}
	return c, nil
}
