// Package api contains HTTP handlers and routing for the GoLedger API.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/vdyabkin01-ux/goledger/internal/user"
)

// UserService is what the API needs from the user package.
// An interface lets tests substitute a fake.
type UserService interface {
	Register(ctx context.Context, email, password string) (user.User, error)
	Authenticate(ctx context.Context, email, password string) (user.User, error)
	Get(ctx context.Context, id string) (user.User, error)
}

// TokenManager issues and verifies access tokens.
type TokenManager interface {
	Issue(userID string) (token string, expiresAt time.Time, err error)
	Parse(token string) (userID string, err error)
}

// Deps are the dependencies of the HTTP API.
type Deps struct {
	Users  UserService
	Tokens TokenManager
	Logger *slog.Logger // optional; slog.Default() if nil
}

type handlers struct {
	users  UserService
	tokens TokenManager
	log    *slog.Logger
}

// NewRouter returns an http.Handler with all API routes registered.
// The "METHOD /path" patterns make the mux answer 405 to other methods.
func NewRouter(d Deps) http.Handler {
	h := &handlers{users: d.Users, tokens: d.Tokens, log: d.Logger}
	if h.log == nil {
		h.log = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthHandler)
	mux.HandleFunc("POST /users", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.Handle("GET /me", h.requireAuth(http.HandlerFunc(h.me)))
	return mux
}
