package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for any token that must not be trusted:
// bad signature, expired, wrong algorithm, malformed and so on.
var ErrInvalidToken = errors.New("invalid token")

const tokenIssuer = "goledger"

// TokenManager issues and verifies signed access tokens (JWT, HS256).
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time // replaced in tests to control time
}

// NewTokenManager creates a TokenManager that signs tokens with secret
// and makes them valid for ttl.
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// Issue creates a token for userID and returns it with its expiry time.
func (m *TokenManager) Issue(userID string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, expiresAt, nil
}

// Parse verifies the token and returns the user ID stored in it.
func (m *TokenManager) Parse(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		// Accept only HS256: blocks "alg: none" and algorithm-switching attacks.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.Subject == "" {
		return "", fmt.Errorf("%w: empty subject", ErrInvalidToken)
	}
	return claims.Subject, nil
}
