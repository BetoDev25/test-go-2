package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"database/sql"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	 "github.com/BetoDev25/test-go-2/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db            *database.Queries
	jwtSecret      string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil && code > 499 {
		log.Printf("Responding with 5XX error: %s", err)
	}

	type errorResponse struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{
		Message: msg,
		Error:   err,
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
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
		db:             dbQueries,
		jwtSecret:      jwtSecret,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))
	mux.Handle("GET /api/healthz", http.HandlerFunc(handlerReadiness))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset",   http.HandlerFunc(apiCfg.handlerReset))
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(apiCfg.handlerValidateChirp))
	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.handlerCreateUser))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handlerCreateChirp))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.handlerGetChirps))
	mux.Handle("GET /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.handlerGetChirp))
	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.handlerLogin))
	mux.Handle("POST /api/refresh", http.HandlerFunc(apiCfg.handlerRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(apiCfg.handlerRevoke))
	mux.Handle("PUT /api/users", http.HandlerFunc(apiCfg.handlerUpdateUser))

	log.Fatal(server.ListenAndServe())
}
