package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxBodyBytes limits request bodies so a client can't send gigabytes.
const maxBodyBytes = 1 << 20 // 1 MiB

// Error codes returned in the "code" field. Tests should assert on these,
// not on human-readable messages.
const (
	codeInvalidJSON        = "invalid_json"
	codeValidation         = "validation_error"
	codeEmailTaken         = "email_taken"
	codeInvalidCredentials = "invalid_credentials"
	codeUnauthorized       = "unauthorized"
	codeInternal           = "internal_error"
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message}})
}

func writeFieldError(w http.ResponseWriter, status int, code, field, message string) {
	writeJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message, Field: field}})
}

// internalError logs the real cause and returns a generic 500,
// so internal details never leak to the client.
func (h *handlers) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, codeInternal, "internal server error")
}

// decodeJSON reads exactly one JSON object into dst.
// Unknown fields and trailing data are rejected.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		switch {
		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")
		case errors.As(err, &maxErr):
			return fmt.Errorf("request body is larger than %d bytes", maxErr.Limit)
		default:
			return fmt.Errorf("malformed JSON: %w", err)
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}
