package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type usersPostReqBody struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := usersPostReqBody{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with decoding the request")
		return
	}
	newUser, err := cfg.dbQueries.CreateUser(r.Context(), requestBody.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong creating a new user in the database")
		return
	}
	localUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}
	respondWithJSON(w, http.StatusCreated, localUser)
}
