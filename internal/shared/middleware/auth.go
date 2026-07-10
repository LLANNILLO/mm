package middleware

import (
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/llannillo/mm/internal/shared/auth"
	"github.com/llannillo/mm/internal/shared/problem"
)

var publicPaths = map[string]bool{
	"POST /users/register": true,
	"GET /health":          true,
}

func Authentication(verifier *oidc.IDTokenVerifier, permSvc auth.PermissionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Method + " " + r.URL.Path
			if publicPaths[key] {
				next.ServeHTTP(w, r)
				return
			}

			rawToken := extractBearerToken(r)
			if rawToken == "" {
				problem.Write(w, problem.Detail{
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
					Detail: "missing or malformed authorization header",
				})
				return
			}

			idToken, err := verifier.Verify(r.Context(), rawToken)
			if err != nil {
				problem.Write(w, problem.Detail{
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
					Detail: "invalid token",
				})
				return
			}

			var rawClaims struct {
				Sub   string `json:"sub"`
				Email string `json:"email"`
			}
			if err := idToken.Claims(&rawClaims); err != nil {
				problem.Write(w, problem.Detail{
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
					Detail: "cannot read token claims",
				})
				return
			}

			userID, perms, err := permSvc.GetUserPermissions(r.Context(), rawClaims.Sub)
			if err != nil {
				problem.Write(w, problem.Detail{
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
					Detail: "cannot resolve user permissions",
				})
				return
			}

			ctx := auth.WithClaims(r.Context(), auth.Claims{
				Sub:         rawClaims.Sub,
				Email:       rawClaims.Email,
				UserID:      userID,
				Permissions: perms,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
