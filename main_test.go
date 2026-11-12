package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("статус = %d, ожидали %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok\n" {
		t.Errorf("тело = %q, ожидали %q", rec.Body.String(), "ok\n")
	}
}
