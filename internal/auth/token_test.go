package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-test-secret-test-secret"

// newTestManager returns a manager whose clock is controlled by *clock.
func newTestManager(clock *time.Time) *TokenManager {
	m := NewTokenManager(testSecret, time.Hour)
	m.now = func() time.Time { return *clock }
	return m
}

func TestToken_IssueAndParse(t *testing.T) {
	clock := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := newTestManager(&clock)

	token, expiresAt, err := m.Issue("user-42")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if want := clock.Add(time.Hour); !expiresAt.Equal(want) {
		t.Errorf("expiresAt = %v, want %v", expiresAt, want)
	}

	userID, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if userID != "user-42" {
		t.Errorf("userID = %q, want user-42", userID)
	}
}

func TestToken_Expiry(t *testing.T) {
	clock := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := newTestManager(&clock)
	token, _, _ := m.Issue("user-42")

	clock = clock.Add(time.Hour - time.Second)
	if _, err := m.Parse(token); err != nil {
		t.Errorf("one second before expiry: unexpected error %v", err)
	}

	clock = clock.Add(2 * time.Second)
	if _, err := m.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("one second after expiry: err = %v, want ErrInvalidToken", err)
	}
}

func TestToken_Rejected(t *testing.T) {
	clock := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := newTestManager(&clock)
	valid, _, _ := m.Issue("user-42")

	otherSecret := NewTokenManager("another-secret-another-secret-12345", time.Hour)
	otherSecret.now = m.now
	foreign, _, _ := otherSecret.Issue("user-42")

	claims := jwt.RegisteredClaims{
		Subject:   "user-42",
		Issuer:    tokenIssuer,
		ExpiresAt: jwt.NewNumericDate(clock.Add(time.Hour)),
	}
	algNone, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)

	wrongIssuerClaims := claims
	wrongIssuerClaims.Issuer = "someone-else"
	wrongIssuer, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, wrongIssuerClaims).SignedString([]byte(testSecret))

	noExpClaims := jwt.RegisteredClaims{Subject: "user-42", Issuer: tokenIssuer}
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, noExpClaims).SignedString([]byte(testSecret))

	noSubjectClaims := claims
	noSubjectClaims.Subject = ""
	noSubject, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, noSubjectClaims).SignedString([]byte(testSecret))

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"garbage", "not.a.jwt"},
		{"signed with another secret", foreign},
		{"tampered signature", valid[:len(valid)-2] + "xx"},
		{"alg none", algNone},
		{"wrong issuer", wrongIssuer},
		{"no expiry", noExp},
		{"no subject", noSubject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := m.Parse(tt.token); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("err = %v, want ErrInvalidToken", err)
			}
		})
	}
}
