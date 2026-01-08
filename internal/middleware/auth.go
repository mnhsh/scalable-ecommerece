package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/herodragmon/scalable-ecommerce/internal/auth"
	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func Auth(cfg *config.APIConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetBearerToken(r.Header)
			if err != nil {
				response.RespondWithError(
					w,
					http.StatusUnauthorized,
					"Missing or invalid token",
					err,
				)
				return
			}

			userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
			if err != nil {
				response.RespondWithError(
					w,
					http.StatusUnauthorized,
					"Invalid or expired token",
					err,
				)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}
