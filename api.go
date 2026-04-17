package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/Eval-99/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hitsHandler(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "text/html; charset=utf-8")
	writter.WriteHeader(200)
	hits := fmt.Sprintf(" <html> <body> <h1>Welcome, Chirpy Admin</h1> <p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	writter.Write([]byte(hits))
}

func (cfg *apiConfig) resetHandler(writter http.ResponseWriter, request *http.Request) {
	if cfg.platform != "dev" {
		writter.WriteHeader(403)
		return
	}

	cfg.db.DeleteAllUsers(context.Background())

	cfg.fileserverHits.Store(0)
	writter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writter.WriteHeader(200)
	writter.Write([]byte("Hits and users have been reset\n"))
}

func (cfg *apiConfig) usersHandler(writter http.ResponseWriter, request *http.Request) {
	req, err := decode(request)
	if err != nil {
		log.Printf("Error decoding request fields: %s", err)
		writter.WriteHeader(500)
		return
	}

	createdUser, err := cfg.db.CreateUser(context.Background(), req.Email)
	if err != nil {
		log.Printf("Error creating createdUser: %s", err)
		writter.WriteHeader(500)
		return
	}

	user := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.Header().Set("Content-Type", "application/json; charset=utf-8")
	writter.WriteHeader(201)
	writter.Write([]byte(dat))
}
