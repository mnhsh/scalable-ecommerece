package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

func GetUserIDFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

func GetUserRoleFromContext(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(userRoleContextKey).(string)
	return role, ok
}

