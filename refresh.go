package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mrmiffmiff/chirpy/internal/auth"
)

type refreshResBody struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	fullRefreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), refresh)
	if err != nil || fullRefreshToken.ExpiresAt.Before(time.Now()) || (fullRefreshToken.RevokedAt.Valid && fullRefreshToken.RevokedAt.Time.Before(time.Now())) {
		respondWithError(w, http.StatusUnauthorized, "Refresh Token does not exist, is expired, or is revoked")
		return
	}
	newJWT, err := auth.MakeJWT(fullRefreshToken.UserID.UUID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Errorf("Something went wrong creating JWT: %w", err).Error())
		return
	}
	respondWithJSON(w, http.StatusOK, refreshResBody{
		Token: newJWT,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = cfg.dbQueries.RevokeToken(r.Context(), refresh)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem revoking refresh token or refresh token does not exist")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
