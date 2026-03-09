package main

import (
	"net/http"
	"database/sql"
	"errors"
	"strings"
	"time"

	auth "github.com/BetoDev25/test-go-2/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	err := errors.New("invalid authorization header")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "cannot verify refresh token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken (r.Context(), parts[1])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 401, "Refresh token does not exist or expired", err)
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get refresh token", err)
			return
		}
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

