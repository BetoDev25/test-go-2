package main

import (
	"net/http"

	"github.com/BetoDev25/test-go-2/internal/database"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps := []database.Chirp{} //this line is not necessary
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError , "Could not get chirps", err)
		return
	}

	response := make([]Chirp, len(dbChirps))
	for index, chirp := range dbChirps {
		response[index] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}
