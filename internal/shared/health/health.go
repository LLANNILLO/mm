package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const checkTimeout = 5 * time.Second

// Checker runs a single health probe.
type Checker interface {
	Check(ctx context.Context) error
}

// CheckResult holds the outcome of one checker.
type CheckResult struct {
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Response is the JSON body written by Handler.
type Response struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}

// Handler runs all registered checkers concurrently and writes the result.
type Handler struct {
	checkers map[string]Checker
}

// NewHandler builds a Handler from a named set of checkers.
func NewHandler(checkers map[string]Checker) *Handler {
	return &Handler{checkers: checkers}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), checkTimeout)
	defer cancel()

	type entry struct {
		name   string
		result CheckResult
	}

	ch := make(chan entry, len(h.checkers))
	for name, c := range h.checkers {
		go func(name string, c Checker) {
			start := time.Now()
			err := c.Check(ctx)
			res := CheckResult{
				Status:     "Healthy",
				DurationMs: time.Since(start).Milliseconds(),
			}
			if err != nil {
				res.Status = "Unhealthy"
				res.Error = err.Error()
			}
			ch <- entry{name, res}
		}(name, c)
	}

	resp := Response{
		Status: "Healthy",
		Checks: make(map[string]CheckResult, len(h.checkers)),
	}
	for range h.checkers {
		e := <-ch
		resp.Checks[e.name] = e.result
		if e.result.Status == "Unhealthy" {
			resp.Status = "Unhealthy"
		}
	}

	status := http.StatusOK
	if resp.Status == "Unhealthy" {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
