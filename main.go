package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/andytrue7/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) writeServerHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
	`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func main() {
	const filePathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	dbQueries:= database.New(db)
	apiCfg.db = dbQueries
	log.Println("Connected to database")

	fileHandler := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileHandler)))

	mux.HandleFunc("GET /api/healthz", handleReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.writeServerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	srv := http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Println("server started at " + port)
	log.Fatal(srv.ListenAndServe())
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		CleanedBody string `json:"cleaned_body"`
	}

	var req parameters
	err := json.NewDecoder(r.Body).Decode(&req)
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

	respondWithJSON(w, http.StatusOK, returnVal{
		CleanedBody: cleaned,
	})
}

func getCleanedChirp(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}