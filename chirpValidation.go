package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

type errorResponseBody struct {
	Error string `json:"error"`
}

type reqBody struct {
	Body string `json:"body"`
}

type validResponseBody struct {
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
	respondWithJSON(w, http.StatusOK, validResponseBody{
		CleanedBody: cleaned,
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody := errorResponseBody{
		Error: msg,
	}
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
