package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ValeriiaGrebneva/Chirpy/internal/auth"
	"github.com/ValeriiaGrebneva/Chirpy/internal/database"
	"github.com/google/uuid"
)

import _ "github.com/lib/pq"

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_ = cfg.fileserverHits.Add(1)
		next.ServeHTTP(resp, req)
	})
}

func (cfg *apiConfig) handlerNRequests(resp http.ResponseWriter, req *http.Request) {
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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirps(resp http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting JWT: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.keyJWT)
	if err != nil {
		log.Printf("Error validating JWT: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 401, respBody)
		return
	}

	type parameters struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"` //we don't need it, since user's ID is found through JWT
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
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
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Chirp is too long",
		}
		responseJSON(resp, 400, respBody)
		return
	}

	cleaned := CleanedBody(params.Body)
	chirp, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{cleaned, userID})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}
	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	responseJSON(resp, 201, respBody)
}

func (cfg *apiConfig) handlerGetChirps(resp http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(req.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	chirpsResponse := make([]Chirp, len(chirps))
	for i, ch := range chirps {
		respBody := Chirp{
			ID:        ch.ID,
			CreatedAt: ch.CreatedAt,
			UpdatedAt: ch.UpdatedAt,
			Body:      ch.Body,
			UserID:    ch.UserID,
		}
		chirpsResponse[i] = respBody
	}
	responseJSON(resp, 200, chirpsResponse)
}

func (cfg *apiConfig) handlerGetChirp(resp http.ResponseWriter, req *http.Request) {
	path := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(path)
	if err != nil {
		log.Printf("Error parsing to UUID: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(req.Context(), chirpUUID)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 404, respBody)
		return
	}

	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	responseJSON(resp, 200, respBody)
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerNewUser(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	if params.Password == "" {
		log.Printf("Password is not provided: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 401, respBody)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	user, err := cfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{sql.NullString{String: params.Email, Valid: true}, hash})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		type returnVals struct {
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

func (cfg *apiConfig) handlerLogin(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(req.Context(), sql.NullString{String: params.Email, Valid: true})
	hashCheck, _ := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || hashCheck == false {
		log.Printf("Error: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 401, respBody)
		return
	}

	refresh, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error making Refresh Token: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	now := time.Now().UTC()
	exp := time.ParseDuration("1440h")
	expiration := now.Add(exp)
	refreshToken, err := cfg.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{refresh, user.ID, expiration})
	if err != nil {
		log.Printf("Error creating Refresh Token: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	signedToken, err := auth.MakeJWT(user.ID, cfg.keyJWT)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	respBody := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email.String,
		Token:        signedToken,
		RefreshToken: refreshToken,
	}
	responseJSON(resp, 200, respBody)
}

func (cfg *apiConfig) handlerRefresh(resp http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Error getting Bearer token: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 401, respBody)
		return
	}

	expiration, err := cfg.dbQueries.GetExpirationFromRefreshToken(req.Context(), refreshToken)
	if err != nil || expiration.Before(time.Now().UTC()) {
		log.Printf("Refresh token doesn't exist or expired")
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 401, respBody)
		return
	}

	userID, err := cfg.dbQueries.GetUserFromRefreshToken(req.Context(), refreshToken)
	//if err != nil

	//should expire in 1 hour - check
	signedToken, err := auth.MakeJWT(user.ID, cfg.keyJWT)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Something went wrong",
		}
		responseJSON(resp, 500, respBody)
		return
	}

	//return code 200 and access token string
}
