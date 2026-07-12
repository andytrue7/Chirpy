package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andytrue7/chirpy/internal/auth"
	"github.com/andytrue7/chirpy/internal/database"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request){
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	var req parameters
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode request")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	isMatch, err := auth.CheckPasswordHash(req.Password, user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check password")
		return
	}

	if !isMatch {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, cfg.tokenExp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken:= auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token: refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email: user.Email,
		},
		Token: token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request){
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	existingToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if existingToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token is revoked")
		return
	}

	if existingToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), existingToken.UserID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err = auth.MakeJWT(user.ID, cfg.jwtSecret, cfg.tokenExp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: token,
	})	
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request){
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
