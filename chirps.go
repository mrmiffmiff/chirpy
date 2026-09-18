package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mrmiffmiff/chirpy/internal/database"
)

type postChirpReqBody struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type chirpResponseBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := postChirpReqBody{}
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

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleaned,
		UserID: uuid.NullUUID{
			Valid: true,
			UUID:  requestBody.UserID,
		},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with posting the new chirp.")
		return
	}
	if !chirp.UserID.Valid {
		respondWithError(w, http.StatusInternalServerError, "Something wrong with User ID posting chirp.")
	}

	resChirp := chirpResponseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	}
	respondWithJSON(w, http.StatusCreated, resChirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong retrieving chirps.")
		return
	}
	resChirps := make([]chirpResponseBody, len(chirps))
	for i, chirp := range chirps {
		resChirps[i] = chirpResponseBody{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID.UUID,
		}
	}
	respondWithJSON(w, http.StatusOK, resChirps)
}

func (cfg *apiConfig) handlerGetSpecificChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	idAsUUID, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID, cannot parse")
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), idAsUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found.")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving chirp")
		return
	}
	resChirp := chirpResponseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	}
	respondWithJSON(w, http.StatusOK, resChirp)
}
