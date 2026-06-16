package problem

import (
	"encoding/json"
	"net/http"
)

// Detail represents an RFC 9457 Problem Details response body.
type Detail struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	Errors any    `json:"errors,omitempty"`
}

// Write writes a Problem Details JSON response.
func Write(w http.ResponseWriter, d Detail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(d.Status)
	json.NewEncoder(w).Encode(d) //nolint:errcheck
}

// WriteInternal writes a generic 500 Problem Details response.
func WriteInternal(w http.ResponseWriter) {
	Write(w, Detail{
		Title:  "Server failure",
		Status: http.StatusInternalServerError,
		Detail: "An unexpected error occurred.",
	})
}
