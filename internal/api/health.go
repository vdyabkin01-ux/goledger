package api

import "net/http"

// healthResponse is the JSON body returned by GET /health.
type healthResponse struct {
	Status string `json:"status"`
}

// HealthHandler reports that the service is up.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
