package api

import (
	"context"
	"net/http"
	"strings"
)

type ctxKeyUserID struct{}

// requireAuth lets the request through only with a valid
// "Authorization: Bearer <token>" header and puts the user ID
// into the request context.
func (h *handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer`)
			writeError(w, http.StatusUnauthorized, codeUnauthorized, "missing or malformed Authorization header")
			return
		}

		userID, err := h.tokens.Parse(token)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, http.StatusUnauthorized, codeUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// bearerToken extracts the token from "Bearer <token>".
func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	token = strings.TrimSpace(token)
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// userIDFromContext returns the user ID set by requireAuth.
func userIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyUserID{}).(string)
	return id
}
