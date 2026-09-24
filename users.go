package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrmiffmiff/chirpy/internal/auth"
	"github.com/mrmiffmiff/chirpy/internal/database"
)

type usersPostReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type usersLoginReqBody struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Expiration int    `json:"expires_in_seconds"`
}

type UserPostResBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type LoginResBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := usersPostReqBody{}
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
	localUser := UserPostResBody{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}
	respondWithJSON(w, http.StatusCreated, localUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	requestBody := usersLoginReqBody{}
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
	var expirationTime time.Duration
	if requestBody.Expiration < 1 || (time.Duration(requestBody.Expiration)*time.Second) > time.Hour {
		expirationTime = 1 * time.Hour
	} else {
		expirationTime = time.Duration(requestBody.Expiration) * time.Second
	}
	jwt, err := auth.MakeJWT(user.ID, cfg.secret, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Something went wrong creating JWT: %w", err).Error())
		return
	}
	localUser := LoginResBody{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     jwt,
	}
	respondWithJSON(w, http.StatusOK, localUser)
}
