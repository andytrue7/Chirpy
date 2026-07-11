package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{
		Error: message,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Println("Error encoding JSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}