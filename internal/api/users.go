package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/vdyabkin01-ux/goledger/internal/user"
)

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// userResponse is the public view of a user: never includes the password hash.
type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type tokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func toUserResponse(u user.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt.UTC()}
}

// register handles POST /users.
func (h *handlers) register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, codeInvalidJSON, err.Error())
		return
	}

	u, err := h.users.Register(r.Context(), req.Email, req.Password)
	var vErr *user.ValidationError
	switch {
	case errors.As(err, &vErr):
		writeFieldError(w, http.StatusBadRequest, codeValidation, vErr.Field, vErr.Message)
	case errors.Is(err, user.ErrEmailTaken):
		writeError(w, http.StatusConflict, codeEmailTaken, "email already registered")
	case err != nil:
		h.internalError(w, r, err)
	default:
		writeJSON(w, http.StatusCreated, toUserResponse(u))
	}
}

// login handles POST /auth/login.
func (h *handlers) login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, codeInvalidJSON, err.Error())
		return
	}

	u, err := h.users.Authenticate(r.Context(), req.Email, req.Password)
	if errors.Is(err, user.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, codeInvalidCredentials, "invalid email or password")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	token, expiresAt, err := h.tokens.Issue(u.ID)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt.UTC(),
	})
}

// me handles GET /me (behind requireAuth).
func (h *handlers) me(w http.ResponseWriter, r *http.Request) {
	u, err := h.users.Get(r.Context(), userIDFromContext(r.Context()))
	if errors.Is(err, user.ErrNotFound) {
		// Token is valid but the user no longer exists.
		writeError(w, http.StatusUnauthorized, codeUnauthorized, "invalid or expired token")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}
