package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"sync"

	"github.com/vdyabkin01-ux/goledger/internal/auth"
)

// Input limits.
const (
	MaxEmailLen    = 254
	MinPasswordLen = 8
	MaxPasswordLen = 72 // bcrypt ignores everything after 72 bytes
)

// dummyHash is compared against when a login email doesn't exist, so a
// wrong email takes as long as a wrong password and response time does
// not reveal which emails are registered.
var dummyHash = sync.OnceValue(func() string {
	hash, _ := auth.HashPassword("dummy-password-for-timing")
	return hash
})

// Service implements registration and login.
type Service struct {
	repo Repository
}

// NewService creates a Service on top of repo.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Register validates input, hashes the password and creates the user.
func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return User{}, err
	}
	return s.repo.Create(ctx, email, hash)
}

// Authenticate returns the user if email and password match,
// otherwise ErrInvalidCredentials.
func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		_ = auth.CheckPassword(dummyHash(), password)
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}

	if err := auth.CheckPassword(u.PasswordHash, password); err != nil {
		if errors.Is(err, auth.ErrPasswordMismatch) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}
	return u, nil
}

// Get returns a user by ID or ErrNotFound.
func (s *Service) Get(ctx context.Context, id string) (User, error) {
	return s.repo.GetByID(ctx, id)
}

// normalizeEmail trims and lowercases the email and checks its format.
func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	invalid := &ValidationError{Field: "email", Message: "must be a valid email address"}

	if email == "" {
		return "", &ValidationError{Field: "email", Message: "is required"}
	}
	if len(email) > MaxEmailLen {
		return "", &ValidationError{Field: "email", Message: "is too long"}
	}
	addr, err := mail.ParseAddress(email)
	// Address must equal the input: rejects forms like "Name <a@b.com>".
	if err != nil || addr.Address != email {
		return "", invalid
	}
	at := strings.LastIndex(email, "@")
	if !strings.Contains(email[at+1:], ".") {
		return "", invalid
	}
	return email, nil
}

func validatePassword(password string) error {
	switch {
	case len(password) < MinPasswordLen:
		return &ValidationError{Field: "password", Message: "must be at least 8 characters"}
	case len(password) > MaxPasswordLen:
		return &ValidationError{Field: "password", Message: "must be at most 72 bytes"}
	}
	return nil
}
