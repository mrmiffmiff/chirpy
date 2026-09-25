package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/mrmiffmiff/chirpy/internal/auth"
)

type polkaWebhookReqBody struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if key != cfg.polka {
		respondWithError(w, http.StatusUnauthorized, "Incorrect API Key")
		return
	}
	decoder := json.NewDecoder(r.Body)
	requestBody := polkaWebhookReqBody{}
	err = decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong with decoding the request")
		return
	}
	if requestBody.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	id, err := uuid.Parse(requestBody.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = cfg.dbQueries.UpgradeUserById(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No such user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
