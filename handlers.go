package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Eval-99/chirpy/internal/database"
	"github.com/google/uuid"
)

type requestFields struct {
	Body     string    `json:"body"`
	Email    string    `json:"email"`
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password"`
}

type responseFields struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	Valid     bool      `json:"valid"`
	Error     string    `json:"error"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func healthHandler(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writter.WriteHeader(200)
	writter.Write([]byte("OK"))
}

func profane(s string) string {
	var cleaned_words []string
	filter := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	words := strings.SplitSeq(s, " ")
	for word := range words {
		if slices.Contains(filter, strings.ToLower(word)) {
			word = "****"
		}
		cleaned_words = append(cleaned_words, word)
	}

	return strings.Join(cleaned_words, " ")
}

func decode(r *http.Request) (requestFields, error) {
	decoder := json.NewDecoder(r.Body)
	req := requestFields{}
	err := decoder.Decode(&req)
	if err != nil {
		return requestFields{}, err
	}
	return req, nil
}

func chirpConvert(chirp database.Chirp) responseFields {
	res := responseFields{}
	res.ID = chirp.ID
	res.CreatedAt = chirp.CreatedAt
	res.UpdatedAt = chirp.UpdatedAt
	res.Body = chirp.Body
	res.UserID = chirp.UserID
	res.Valid = true

	return res
}
