package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ValeriiaGrebneva/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

import _ "github.com/lib/pq"

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platformAPI    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_ = cfg.fileserverHits.Add(1)
		next.ServeHTTP(resp, req)
	})
}

func (cfg *apiConfig) handlerNRequests(resp http.ResponseWriter, req *http.Request) {
	//resp.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
	resp.Header().Set("Content-Type", "text/html")
	resp.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetRequests(resp http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	if cfg.platformAPI != "dev" {
		resp.WriteHeader(403)
		return
	}
	err := cfg.dbQueries.ResetUsers(req.Context())
	if err != nil {
		log.Printf("Error resetting users: %s", err)
	}
}

func handlerFunc(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}

func responseJSON(resp http.ResponseWriter, code int, response interface{}) {
	resp.WriteHeader(code)
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		resp.WriteHeader(500)
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(dat)
	return
}

func CleanedBody(msg string) string {
	splitted := strings.Split(msg, " ")
	for i, word := range splitted {
		word = strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			splitted[i] = "****"
		}
	}
	return strings.Join(splitted, " ")
}

func handlerChirp(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	if len(params.Body) > 140 {
		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Chirp is too long",
		}
		responseJSON(resp, 400, respBody)
		return
	}

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Cleaned_body string `json:"cleaned_body"`
	}
	cleaned := CleanedBody(params.Body)
	respBody := returnVals{
		Cleaned_body: cleaned,
	}
	responseJSON(resp, 200, respBody)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerNewUser(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	user, err := cfg.dbQueries.CreateUser(req.Context(), sql.NullString{String: params.Email, Valid: true})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}
	respBody := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email.String,
	}
	responseJSON(resp, 201, respBody)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
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
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", handlerFunc)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerNRequests)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerResetRequests)
	serveMux.HandleFunc("POST /api/validate_chirp", handlerChirp)
	serveMux.HandleFunc("POST /api/users", apiCfg.handlerNewUser)
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverStruct := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err = serverStruct.ListenAndServe()
	fmt.Println(err)
}
