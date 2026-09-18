package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

type reqBody struct {
	Body string `json:"body"`
}

type validVerificationResponseBody struct {
	CleanedBody string `json:"cleaned_body"`
}

func handlerChirpValidation(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := reqBody{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with decoding the request")
		return
	}
	if len(requestBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(requestBody.Body, " ")
	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	respondWithJSON(w, http.StatusOK, validVerificationResponseBody{
		CleanedBody: cleaned,
	})
}
