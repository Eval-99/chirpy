package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Eval-99/chirpy/internal/auth"
	"github.com/Eval-99/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
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

	cfg.db.DeleteAllUsers(request.Context())

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

	if req.Email == "" || req.Password == "" {
		log.Printf("Error creating user, Email or Password missing: %s", err)
		writter.WriteHeader(500)
		return
	}

	pass, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		writter.WriteHeader(500)
		return
	}

	params := database.CreateUserParams{Email: req.Email, HashedPassword: pass}
	createdUser, err := cfg.db.CreateUser(request.Context(), params)
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

func (cfg *apiConfig) loginHandler(writter http.ResponseWriter, request *http.Request) {
	req, err := decode(request)
	if req.Email == "" || req.Password == "" {
		log.Printf("Error looking up user, Email or Password missing: %s", err)
		writter.WriteHeader(500)
		return
	}

	dbUser, err := cfg.db.UsersByEmail(request.Context(), req.Email)
	if err != nil {
		log.Printf("Incorrect email or password")
		writter.WriteHeader(401)
		return
	}

	isValid, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
	if err != nil || !isValid {
		log.Printf("Incorrect email or password")
		writter.WriteHeader(401)
		return
	}

	if req.ExpiresInSeconds == 0 || req.ExpiresInSeconds > 3600 {
		req.ExpiresInSeconds = 3600
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, time.Second*time.Duration(req.ExpiresInSeconds))
	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		writter.WriteHeader(500)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     token,
	}

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.Header().Set("Content-Type", "application/json; charset=utf-8")
	writter.WriteHeader(200)
	writter.Write([]byte(dat))
}

func (cfg *apiConfig) chirpsHandler(writter http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		log.Printf("Error token is missing or malformed: %s", err)
		writter.WriteHeader(401)
		return
	}

	validatedUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Error token is invalid: %s", err)
		writter.WriteHeader(401)
		return
	}

	req, err := decode(request)
	if err != nil {
		log.Printf("Error decoding request fields: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(req.Body) > 140 {
		res := responseFields{}
		res.Error = "Chirp is too long"

		dat, err := json.Marshal(res)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			writter.WriteHeader(500)
			return
		}

		writter.WriteHeader(400)
		writter.Write([]byte(dat))
		return
	}

	chirp, err := cfg.db.CreateChirp(request.Context(), database.CreateChirpParams{UserID: validatedUserID, Body: profane(req.Body)})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		writter.WriteHeader(500)
		return
	}

	res := chirpConvert(chirp)

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.WriteHeader(201)
	writter.Write([]byte(dat))
}

func (cfg *apiConfig) allChirpsHandler(writter http.ResponseWriter, request *http.Request) {
	dbChirps, err := cfg.db.AllChirps(request.Context())
	if err != nil {
		log.Printf("Error fetching all chirps: %s", err)
		writter.WriteHeader(500)
		return
	}

	chirpSlice := []responseFields{}
	for _, chirp := range dbChirps {
		res := chirpConvert(chirp)
		chirpSlice = append(chirpSlice, res)
	}

	dat, err := json.Marshal(chirpSlice)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.Header().Set("Content-Type", "application/json; charset=utf-8")
	writter.WriteHeader(200)
	writter.Write([]byte(dat))
}

func (cfg *apiConfig) chirpsIDHandler(writter http.ResponseWriter, request *http.Request) {
	chirpID, err := uuid.Parse(request.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing chirp ID, not a valid uuid: %s", err)
		writter.WriteHeader(404)
		return
	}

	chirp, err := cfg.db.ChirpsID(request.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirp ID, not in database: %s", err)
		writter.WriteHeader(404)
		return
	}

	res := chirpConvert(chirp)

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.Header().Set("Content-Type", "application/json; charset=utf-8")
	writter.WriteHeader(200)
	writter.Write([]byte(dat))
}
