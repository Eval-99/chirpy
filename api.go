package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hitsHandler(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "text/html; charset=utf-8")
	writter.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf(" <html> <body> <h1>Welcome, Chirpy Admin</h1> <p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	writter.Write([]byte(hits))
}

func (cfg *apiConfig) resetHitsHandler(writter http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
	writter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writter.WriteHeader(http.StatusOK)
	writter.Write([]byte("Hits have been reset\n"))
}
