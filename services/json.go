package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Largest JSON request body accepted. Requests here are a query or a password, so this is generous.
const maxJSONBodyBytes = 1 << 20

// decodeJSON reads a JSON request body into v
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)).Decode(v)
}

// writeJSON writes v as a JSON response with the given status
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("Error writing json", "err", err)
	}
}
