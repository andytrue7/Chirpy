package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/andytrue7/chirpy/internal/auth"
	"github.com/andytrue7/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct{
	ID uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func(cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request){
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req parameters
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode request")
		return
	}

	const maxChirpLength = 140
	if len(req.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}

	cleaned:= getCleanedChirp(req.Body, badWords)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userID,
		Body: cleaned,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID: chirp.ID,
			UserID: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
		},
	})
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request){
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps")
		return
	}

	chirpsToRespond:= []Chirp{}
	for _, chirp:= range chirps{
		chirpsToRespond = append(chirpsToRespond, Chirp{
			ID: chirp.ID,
			UserID: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsToRespond)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request){
	type response struct {
		Chirp
	}

	idParam := r.PathValue("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			ID: chirp.ID,
			UserID: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
		},
	})
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request){
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpID := r.PathValue("id")

	uuidChirpId, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	
	chirp, err := cfg.db.GetChirpByID(r.Context(), uuidChirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), uuidChirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}