package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(hash, "correct horse") {
		t.Fatal("hash contains the plain password")
	}
	if err := CheckPassword(hash, "correct horse battery staple"); err != nil {
		t.Errorf("CheckPassword with right password: %v", err)
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	err = CheckPassword(hash, "Secret-password")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("err = %v, want ErrPasswordMismatch", err)
	}
}

func TestHashPassword_IsSalted(t *testing.T) {
	a, _ := HashPassword("same-password")
	b, _ := HashPassword("same-password")
	if a == b {
		t.Error("two hashes of the same password are equal; salt is missing")
	}
}

func TestCheckPassword_BrokenHash(t *testing.T) {
	err := CheckPassword("not-a-bcrypt-hash", "whatever")
	if err == nil || errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("err = %v, want a non-mismatch error", err)
	}
}
