package middleware

import (
	"context"
	"net/http"
	"strings"

	"recipes/pkg/auth"
	"recipes/pkg/response"
)

type principalContextKey string

const principalKey principalContextKey = "recipes.principal"

// RequireAuth ensures a valid bearer token is present.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			response.Unauthorized(w, "Missing or invalid Authorization header")
			return
		}

		principal, err := auth.ParseAccessToken(token)
		if err != nil {
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), principalKey, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole ensures caller has one of the allowed roles.
func RequireRole(allowedRoles ...string) Middleware {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, "Authentication required")
				return
			}

			if _, roleAllowed := allowed[principal.Role]; !roleAllowed {
				response.Forbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSelfOrAdmin ensures user matches path param unless role is admin.
func RequireSelfOrAdmin(pathParam string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, "Authentication required")
				return
			}

			targetUserID := r.PathValue(pathParam)
			if targetUserID == "" {
				response.BadRequest(w, "User ID is required")
				return
			}

			if principal.Role == "admin" || principal.UserID == targetUserID {
				next.ServeHTTP(w, r)
				return
			}

			response.Forbidden(w, "You can only access your own resources")
		})
	}
}

// PrincipalFromContext returns authenticated principal details.
func PrincipalFromContext(ctx context.Context) (auth.Principal, bool) {
	principal, ok := ctx.Value(principalKey).(auth.Principal)
	return principal, ok
}

func bearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
