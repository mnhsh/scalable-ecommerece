package middleware

import (
	"net/http"

	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(userRoleContextKey).(string)
		if !ok || role != "admin" {
			response.RespondWithError(
				w,
				http.StatusForbidden,
				"admin access required",
				nil,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

