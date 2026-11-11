// Package api contains HTTP handlers and routing for the GoLedger API.
package api

import (
	"encoding/json"
	"net/http"
)

// healthResponse is the JSON body returned by GET /health.
type healthResponse struct {
	Status string `json:"status"`
}

// NewRouter returns an http.Handler with all API routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// The "GET " prefix makes the mux reply 405 Method Not Allowed to other methods.
	mux.HandleFunc("GET /health", HealthHandler)
	return mux
}

// HealthHandler reports that the service is up.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
