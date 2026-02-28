package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/noggrj/autorepair/internal/platform/auth"
	"github.com/noggrj/autorepair/internal/platform/errors"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.Unauthorized(w, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.Unauthorized(w, "invalid authorization header format")
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			errors.Unauthorized(w, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok {
				errors.Unauthorized(w, "user not authenticated")
				return
			}

			if claims.Role != role && claims.Role != "admin" { // Admin can access everything
				errors.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
