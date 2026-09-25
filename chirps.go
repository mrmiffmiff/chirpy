package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mrmiffmiff/chirpy/internal/auth"
	"github.com/mrmiffmiff/chirpy/internal/database"
)

type postChirpReqBody struct {
	Body string `json:"body"`
}

type chirpResponseBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
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
	if chirp.UserID.UUID != userid {
		respondWithError(w, http.StatusForbidden, "User is not author of the chirp")
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem deleting chirp")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	decoder := json.NewDecoder(r.Body)
	requestBody := postChirpReqBody{}
	err = decoder.Decode(&requestBody)
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
			UUID:  userid,
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
	authorID := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if authorID == "" {
		chirps, err = cfg.dbQueries.GetChirps(r.Context())
	} else {
		authorIDasUUID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Something went wrong parsing Author ID in query")
			return
		}
		chirps, err = cfg.dbQueries.GetChirpsByAuthor(r.Context(), uuid.NullUUID{
			UUID:  authorIDasUUID,
			Valid: true,
		})
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong retrieving chirps.")
		return
	}
	sortDirection := r.URL.Query().Get("sort")
	if sortDirection == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
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
