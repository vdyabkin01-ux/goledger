package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vdyabkin01-ux/goledger/internal/user"
	"github.com/vdyabkin01-ux/goledger/internal/user/usertest"
)

func newService() *user.Service {
	return user.NewService(usertest.NewMemoryRepo())
}

func TestRegister_Success(t *testing.T) {
	svc := newService()

	u, err := svc.Register(context.Background(), "  Vlad@Example.COM ", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.ID == "" {
		t.Error("ID is empty")
	}
	if u.Email != "vlad@example.com" {
		t.Errorf("Email = %q, want normalized vlad@example.com", u.Email)
	}
	if u.PasswordHash == "" || u.PasswordHash == "password123" {
		t.Errorf("PasswordHash = %q, want a bcrypt hash", u.PasswordHash)
	}
}

func TestRegister_DuplicateEmailIgnoresCase(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "vlad@example.com", "password123"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	_, err := svc.Register(ctx, "VLAD@example.com", "another-password")
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Errorf("err = %v, want ErrEmailTaken", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	longLocal := strings.Repeat("a", user.MaxEmailLen) + "@example.com"

	tests := []struct {
		name      string
		email     string
		password  string
		wantField string // "" means registration must succeed
	}{
		{"empty email", "", "password123", "email"},
		{"spaces only email", "   ", "password123", "email"},
		{"no at sign", "vlad.example.com", "password123", "email"},
		{"no domain dot", "vlad@localhost", "password123", "email"},
		{"display name form", "Vlad <vlad@example.com>", "password123", "email"},
		{"email too long", longLocal, "password123", "email"},
		{"password 7 chars", "a@example.com", "1234567", "password"},
		{"password 8 chars ok", "b@example.com", "12345678", ""},
		{"password 72 bytes ok", "c@example.com", strings.Repeat("p", 72), ""},
		{"password 73 bytes", "d@example.com", strings.Repeat("p", 73), "password"},
		{"empty password", "e@example.com", "", "password"},
	}

	svc := newService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(context.Background(), tt.email, tt.password)

			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var vErr *user.ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("err = %v, want *ValidationError", err)
			}
			if vErr.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", vErr.Field, tt.wantField)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	registered, err := svc.Register(ctx, "vlad@example.com", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{"correct credentials", "vlad@example.com", "password123", nil},
		{"email in other case", " VLAD@Example.com", "password123", nil},
		{"wrong password", "vlad@example.com", "password124", user.ErrInvalidCredentials},
		{"password in other case", "vlad@example.com", "PASSWORD123", user.ErrInvalidCredentials},
		{"unknown email", "nobody@example.com", "password123", user.ErrInvalidCredentials},
		{"empty password", "vlad@example.com", "", user.ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.Authenticate(ctx, tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && u.ID != registered.ID {
				t.Errorf("ID = %q, want %q", u.ID, registered.ID)
			}
		})
	}
}

func TestGet(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	registered, _ := svc.Register(ctx, "vlad@example.com", "password123")

	got, err := svc.Get(ctx, registered.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Email != "vlad@example.com" {
		t.Errorf("Email = %q", got.Email)
	}

	if _, err := svc.Get(ctx, "no-such-id"); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
