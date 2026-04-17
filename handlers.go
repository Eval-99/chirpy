package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func healthHandler(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writter.WriteHeader(http.StatusOK)
	writter.Write([]byte("OK"))
}

func ValidateChirpHandler(writter http.ResponseWriter, request *http.Request) {
	type requestFields struct {
		Body string `json:"body"`
	}

	type responseFields struct {
		Valid     bool   `json:"valid"`
		Error     string `json:"error"`
		BodyClean string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(request.Body)
	req := requestFields{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request fields: %s", err)
		writter.WriteHeader(500)
		return
	}

	res := responseFields{}
	writter.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(req.Body) > 140 {
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

	res.Valid = true
	res.BodyClean = profane(req.Body)

	dat, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writter.WriteHeader(500)
		return
	}

	writter.WriteHeader(200)
	writter.Write([]byte(dat))
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
