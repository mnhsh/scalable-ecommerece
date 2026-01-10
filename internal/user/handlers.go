package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/herodragmon/scalable-ecommerce/internal/auth"
	"github.com/herodragmon/scalable-ecommerce/internal/config"
	"github.com/herodragmon/scalable-ecommerce/internal/database"
	"github.com/herodragmon/scalable-ecommerce/internal/middleware"
	"github.com/herodragmon/scalable-ecommerce/internal/response"
)

func handlerLogin(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type userResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	type resp struct {
		User  userResponse `json:"user"`
		Token string       `json:"token"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		response.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	accessToken, err := auth.MakeJWT(
    user.ID,
    "user",          // role
    cfg.JWTSecret,
    time.Hour,
	)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = cfg.DB.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour * 60),
		},
	)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   cfg.Platform != "dev",
		Path:     "/api",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour * 60),
	})

	response.RespondWithJSON(w, http.StatusOK, resp{
		User: userResponse{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token: accessToken,
	})
}

func handlerMe(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		response.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := cfg.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "User not found", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, user)
}

func handlerUsersCreate(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	if params.Password == "" || params.Email == "" {
		response.RespondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, user)
}

func handlerRefresh(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	type resp struct {
		Token string `json:"token"`
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Missing refresh token", err)
		return
	}

	refreshToken := cookie.Value

	user, err := cfg.DB.GetUserByRefreshToken(r.Context(), refreshToken)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(
    user.ID,
    "user",          // role
    cfg.JWTSecret,
    time.Hour,
	)
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	response.RespondWithJSON(w, http.StatusOK, resp{
		Token: accessToken,
	})
}

func handlerRevoke(cfg *config.APIConfig, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.RespondWithError(w, http.StatusUnauthorized, "Missing refresh token", err)
		return
	}

	refreshToken := cookie.Value

	if err := cfg.DB.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api",
		HttpOnly: true,
		Secure:   cfg.Platform != "dev",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})

	w.WriteHeader(http.StatusNoContent)
}
