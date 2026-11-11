package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vdyabkin01-ux/goledger/internal/auth"
	"github.com/vdyabkin01-ux/goledger/internal/user"
	"github.com/vdyabkin01-ux/goledger/internal/user/usertest"
)

const testJWTSecret = "api-test-secret-api-test-secret-123"

// testAPI wires the real user service and token manager with an
// in-memory repository: everything except the database is real.
type testAPI struct {
	router http.Handler
	tokens *auth.TokenManager
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	tokens := auth.NewTokenManager(testJWTSecret, time.Hour)
	router := NewRouter(Deps{
		Users:  user.NewService(usertest.NewMemoryRepo()),
		Tokens: tokens,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	return &testAPI{router: router, tokens: tokens}
}

// do sends a request and returns the recorded response.
func (a *testAPI) do(t *testing.T, method, path, body string, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return v
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string) errorDetail {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, wantStatus, rec.Body.String())
	}
	body := decode[errorResponse](t, rec)
	if body.Error.Code != wantCode {
		t.Errorf("error code = %q, want %q", body.Error.Code, wantCode)
	}
	return body.Error
}

func TestRegister_Created(t *testing.T) {
	a := newTestAPI(t)

	rec := a.do(t, http.MethodPost, "/users", `{"email":"Vlad@Example.com","password":"password123"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response leaks password data: %s", rec.Body.String())
	}
	u := decode[userResponse](t, rec)
	if u.ID == "" || u.Email != "vlad@example.com" || u.CreatedAt.IsZero() {
		t.Errorf("unexpected user: %+v", u)
	}
}

func TestRegister_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
		wantField  string
	}{
		{"empty body", ``, 400, codeInvalidJSON, ""},
		{"not json", `email=a@b.com`, 400, codeInvalidJSON, ""},
		{"json array", `[]`, 400, codeInvalidJSON, ""},
		{"unknown field", `{"email":"a@example.com","password":"password123","admin":true}`, 400, codeInvalidJSON, ""},
		{"two objects", `{"email":"a@example.com","password":"password123"}{}`, 400, codeInvalidJSON, ""},
		{"wrong type", `{"email":123,"password":"password123"}`, 400, codeInvalidJSON, ""},
		{"invalid email", `{"email":"not-an-email","password":"password123"}`, 400, codeValidation, "email"},
		{"short password", `{"email":"a@example.com","password":"short"}`, 400, codeValidation, "password"},
		{"missing password", `{"email":"a@example.com"}`, 400, codeValidation, "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAPI(t)
			rec := a.do(t, http.MethodPost, "/users", tt.body)
			detail := assertError(t, rec, tt.wantStatus, tt.wantCode)
			if detail.Field != tt.wantField {
				t.Errorf("field = %q, want %q", detail.Field, tt.wantField)
			}
		})
	}
}

func TestRegister_BodyTooLarge(t *testing.T) {
	a := newTestAPI(t)
	huge := `{"email":"a@example.com","password":"` + strings.Repeat("x", maxBodyBytes) + `"}`

	rec := a.do(t, http.MethodPost, "/users", huge)
	assertError(t, rec, http.StatusBadRequest, codeInvalidJSON)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	a := newTestAPI(t)
	body := `{"email":"vlad@example.com","password":"password123"}`

	if rec := a.do(t, http.MethodPost, "/users", body); rec.Code != http.StatusCreated {
		t.Fatalf("first register: status %d", rec.Code)
	}
	rec := a.do(t, http.MethodPost, "/users", `{"email":"VLAD@example.com","password":"other-password"}`)
	assertError(t, rec, http.StatusConflict, codeEmailTaken)
}

func TestLogin(t *testing.T) {
	a := newTestAPI(t)
	a.do(t, http.MethodPost, "/users", `{"email":"vlad@example.com","password":"password123"}`)

	t.Run("success returns a working token", func(t *testing.T) {
		rec := a.do(t, http.MethodPost, "/auth/login", `{"email":"vlad@example.com","password":"password123"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d; body: %s", rec.Code, rec.Body.String())
		}
		tok := decode[tokenResponse](t, rec)
		if tok.TokenType != "Bearer" || tok.AccessToken == "" {
			t.Fatalf("unexpected token response: %+v", tok)
		}
		if !tok.ExpiresAt.After(time.Now()) {
			t.Errorf("expires_at %v is not in the future", tok.ExpiresAt)
		}
		if _, err := a.tokens.Parse(tok.AccessToken); err != nil {
			t.Errorf("issued token does not parse: %v", err)
		}
	})

	t.Run("wrong password and unknown email look the same", func(t *testing.T) {
		wrongPass := a.do(t, http.MethodPost, "/auth/login", `{"email":"vlad@example.com","password":"nope-nope"}`)
		unknown := a.do(t, http.MethodPost, "/auth/login", `{"email":"ghost@example.com","password":"nope-nope"}`)

		assertError(t, wrongPass, http.StatusUnauthorized, codeInvalidCredentials)
		assertError(t, unknown, http.StatusUnauthorized, codeInvalidCredentials)
		if wrongPass.Body.String() != unknown.Body.String() {
			t.Errorf("responses differ, which reveals registered emails:\n%s\n%s",
				wrongPass.Body.String(), unknown.Body.String())
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		rec := a.do(t, http.MethodPost, "/auth/login", `{"email":`)
		assertError(t, rec, http.StatusBadRequest, codeInvalidJSON)
	})
}

func TestMe(t *testing.T) {
	a := newTestAPI(t)
	created := decode[userResponse](t,
		a.do(t, http.MethodPost, "/users", `{"email":"vlad@example.com","password":"password123"}`))
	tok := decode[tokenResponse](t,
		a.do(t, http.MethodPost, "/auth/login", `{"email":"vlad@example.com","password":"password123"}`))

	t.Run("valid token", func(t *testing.T) {
		rec := a.do(t, http.MethodGet, "/me", "", "Authorization", "Bearer "+tok.AccessToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d; body: %s", rec.Code, rec.Body.String())
		}
		me := decode[userResponse](t, rec)
		if me.ID != created.ID || me.Email != "vlad@example.com" {
			t.Errorf("me = %+v, want user %s", me, created.ID)
		}
	})

	t.Run("scheme is case-insensitive", func(t *testing.T) {
		rec := a.do(t, http.MethodGet, "/me", "", "Authorization", "bearer "+tok.AccessToken)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	tokenForGhost, _, _ := a.tokens.Issue("11111111-1111-1111-1111-111111111111")

	unauthorized := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"basic scheme", "Basic dmxhZDpwYXNzd29yZA=="},
		{"bearer without token", "Bearer "},
		{"token without scheme", tok.AccessToken},
		{"garbage token", "Bearer abc.def.ghi"},
		{"tampered token", "Bearer " + tok.AccessToken + "x"},
		{"token of deleted user", "Bearer " + tokenForGhost},
	}
	for _, tt := range unauthorized {
		t.Run(tt.name, func(t *testing.T) {
			var rec *httptest.ResponseRecorder
			if tt.header == "" {
				rec = a.do(t, http.MethodGet, "/me", "")
			} else {
				rec = a.do(t, http.MethodGet, "/me", "", "Authorization", tt.header)
			}
			assertError(t, rec, http.StatusUnauthorized, codeUnauthorized)
			if rec.Header().Get("WWW-Authenticate") == "" && tt.name != "token of deleted user" {
				t.Error("WWW-Authenticate header is missing")
			}
		})
	}
}

// failingUsers is a UserService whose every call fails like a broken database.
type failingUsers struct{}

var errDBDown = errors.New("connection refused: secret-host:5432")

func (failingUsers) Register(context.Context, string, string) (user.User, error) {
	return user.User{}, errDBDown
}
func (failingUsers) Authenticate(context.Context, string, string) (user.User, error) {
	return user.User{}, errDBDown
}
func (failingUsers) Get(context.Context, string) (user.User, error) {
	return user.User{}, errDBDown
}

func TestInternalErrorsDoNotLeakDetails(t *testing.T) {
	tokens := auth.NewTokenManager(testJWTSecret, time.Hour)
	router := NewRouter(Deps{
		Users:  failingUsers{},
		Tokens: tokens,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	token, _, _ := tokens.Issue("some-user")

	requests := []*http.Request{
		httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email":"a@example.com","password":"password123"}`)),
		httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"a@example.com","password":"password123"}`)),
		httptest.NewRequest(http.MethodGet, "/me", nil),
	}
	requests[2].Header.Set("Authorization", "Bearer "+token)

	for _, req := range requests {
		t.Run(req.Method+" "+req.URL.Path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assertError(t, rec, http.StatusInternalServerError, codeInternal)
			if strings.Contains(rec.Body.String(), "secret-host") {
				t.Errorf("response leaks internal error: %s", rec.Body.String())
			}
		})
	}
}
