package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	 "github.com/BetoDev25/test-go-2/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:   ":8080",
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		database:       dbQueries,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))
	mux.Handle("GET /api/healthz", http.HandlerFunc(handlerReadiness))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset",   http.HandlerFunc(apiCfg.handlerReset))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(apiCfg.handlerValidateChirp))

	log.Fatal(server.ListenAndServe())
}
