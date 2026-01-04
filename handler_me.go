package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

