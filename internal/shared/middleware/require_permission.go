package middleware

import (
	"net/http"
	"slices"

	"github.com/llannillo/mm/internal/shared/auth"
	"github.com/llannillo/mm/internal/shared/problem"
)

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok || !slices.Contains(claims.Permissions, permission) {
				problem.Write(w, problem.Detail{
					Title:  "Forbidden",
					Status: http.StatusForbidden,
					Detail: "insufficient permissions",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
