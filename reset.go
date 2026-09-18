package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Endpoint disallowed outside of development environment.")
		return
	}
	cfg.fileserverHits.Swap(0)
	cfg.dbQueries.ClearUsers(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0\n"))
}
