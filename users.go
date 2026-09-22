package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrmiffmiff/chirpy/internal/auth"
	"github.com/mrmiffmiff/chirpy/internal/database"
)

type usersPostOrLoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := usersPostOrLoginReqBody{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with decoding the request")
		return
	}
	hashedPass, err := auth.HashPassword(requestBody.Password)
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with hashing the password")
		return
	}
	parms := database.CreateUserParams{
		Email:          requestBody.Email,
		HashedPassword: hashedPass,
	}
	newUser, err := cfg.dbQueries.CreateUser(r.Context(), parms)
	if err != nil {
		log.Println(err)
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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := usersPostOrLoginReqBody{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with decoding the request")
		return
	}
	user, err := cfg.dbQueries.GetUserByEmailAddress(r.Context(), requestBody.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	match, err := auth.CheckPasswordHash(requestBody.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	localUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusOK, localUser)
}
