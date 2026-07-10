package main

import (
	"log"
	"net/http"
)

func main() {
	const filePathRoot = "."
	const port = "8080"
	mux := http.NewServeMux()
	
	fileHandler := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("/app/", http.StripPrefix("/app", fileHandler))

	mux.HandleFunc("/healthz", handleReadiness)

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