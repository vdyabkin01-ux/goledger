// Package user contains the user model and registration/login logic.
package user

import (
	"context"
	"errors"
	"time"
)

// User is a registered account.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

var (
	// ErrEmailTaken means a user with this email already exists.
	ErrEmailTaken = errors.New("email already registered")
	// ErrNotFound means no user matches the query.
	ErrNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned for a wrong email or password.
	// It deliberately doesn't say which one was wrong.
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// ValidationError describes invalid client input for one field.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Repository stores users. Implementations: Postgres in production,
// an in-memory fake in unit tests (package usertest).
type Repository interface {
	// Create saves a new user. Returns ErrEmailTaken if the email is used.
	Create(ctx context.Context, email, passwordHash string) (User, error)
	// GetByEmail returns ErrNotFound if there is no such user.
	GetByEmail(ctx context.Context, email string) (User, error)
	// GetByID returns ErrNotFound if there is no such user.
	GetByID(ctx context.Context, id string) (User, error)
}
