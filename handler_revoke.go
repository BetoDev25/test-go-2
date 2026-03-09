package main

import (
	"net/http"
	"database/sql"
	"errors"
	"strings"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	err := errors.New("invalid authorization header")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "cannot verify refresh token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), parts[1])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 401, "Refresh token does not exist or expired", err)
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get refresh token", err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)

}
