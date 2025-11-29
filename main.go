package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ValeriiaGrebneva/Chirpy/internal/database"
	"github.com/joho/godotenv"
)

import _ "github.com/lib/pq"

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platformAPI    string
	keyJWT         string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	key := os.Getenv("KEY_JWT")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
	}
	dbQueriesNew := database.New(db)

	var counter atomic.Int32
	counter.Store(0)
	apiCfg := apiConfig{
		fileserverHits: counter,
		dbQueries:      dbQueriesNew,
		platformAPI:    platform,
		keyJWT:         key,
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", handlerFunc)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerNRequests)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerResetRequests)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	serveMux.HandleFunc("POST /api/users", apiCfg.handlerNewUser)
	serveMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverStruct := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err = serverStruct.ListenAndServe()
	fmt.Println(err)
}
