package middleware

import (
	"log/slog"
	"net/http"

	"github.com/llannillo/mm/internal/shared/problem"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(r.Context(), "Panic recovered",
						"panic", rec,
						"method", r.Method,
						"path", r.URL.Path,
					)
					problem.WriteInternal(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
