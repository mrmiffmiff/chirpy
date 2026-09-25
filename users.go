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

type usersPutChangeEmailAndPasswordBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type usersLoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserPostOrPutResBody struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type LoginResBody struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	Token       string    `json:"token"`
	Refresh     string    `json:"refresh_token"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
	localUser := UserPostOrPutResBody{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed.Valid && newUser.IsChirpyRed.Bool,
	}
	respondWithJSON(w, http.StatusCreated, localUser)
}

func (cfg *apiConfig) handlerPutUsers(w http.ResponseWriter, r *http.Request) {
	access, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userId, err := auth.ValidateJWT(access, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	decoder := json.NewDecoder(r.Body)
	requestBody := usersPutChangeEmailAndPasswordBody{}
	err = decoder.Decode(&requestBody)
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
	user, err := cfg.dbQueries.UpdateUserEmailAndPassword(r.Context(), database.UpdateUserEmailAndPasswordParams{
		ID:             userId,
		Email:          requestBody.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong updating user in database")
		return
	}
	localUser := UserPostOrPutResBody{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Valid && user.IsChirpyRed.Bool,
	}
	respondWithJSON(w, http.StatusOK, localUser)
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
	jwt, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Something went wrong creating JWT: %w", err).Error())
		return
	}
	refresh, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: auth.MakeRefreshToken(),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Problem inserting new refresh token into database: %w", err).Error())
	}
	localUser := LoginResBody{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		Token:       jwt,
		Refresh:     refresh,
		IsChirpyRed: user.IsChirpyRed.Valid && user.IsChirpyRed.Bool,
	}
	respondWithJSON(w, http.StatusOK, localUser)
}
