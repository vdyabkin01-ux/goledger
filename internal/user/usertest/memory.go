// Package usertest provides an in-memory user.Repository for unit tests.
package usertest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vdyabkin01-ux/goledger/internal/user"
)

// MemoryRepo is a thread-safe in-memory user.Repository.
type MemoryRepo struct {
	mu     sync.Mutex
	users  map[string]user.User // by ID
	nextID int
}

// NewMemoryRepo returns an empty repository.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{users: make(map[string]user.User)}
}

// Create implements user.Repository.
func (r *MemoryRepo) Create(_ context.Context, email, passwordHash string) (user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == email {
			return user.User{}, user.ErrEmailTaken
		}
	}
	r.nextID++
	u := user.User{
		ID:           fmt.Sprintf("00000000-0000-0000-0000-%012d", r.nextID),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	r.users[u.ID] = u
	return u, nil
}

// GetByEmail implements user.Repository.
func (r *MemoryRepo) GetByEmail(_ context.Context, email string) (user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return user.User{}, user.ErrNotFound
}

// GetByID implements user.Repository.
func (r *MemoryRepo) GetByID(_ context.Context, id string) (user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}
