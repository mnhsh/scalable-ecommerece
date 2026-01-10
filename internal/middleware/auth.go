package middleware

import (
	"context"
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/auth"
	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

type contextKey string

const (
	userIDContextKey   contextKey = "userID"
	userRoleContextKey contextKey = "userRole"
)

func Auth(cfg *config.APIConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetBearerToken(r.Header)
			if err != nil {
				response.RespondWithError(
					w,
					http.StatusUnauthorized,
					"missing or invalid token",
					err,
				)
				return
			}

			userID, role, err := auth.ValidateJWT(token, cfg.JWTSecret)
			if err != nil {
				response.RespondWithError(
					w,
					http.StatusUnauthorized,
					"invalid or expired token",
					err,
				)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			ctx = context.WithValue(ctx, userRoleContextKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

