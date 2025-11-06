package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	for i, word := range(splitted) {
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

func main() {
	var counter atomic.Int32
	counter.Store(0)
	apiCfg := apiConfig{
		fileserverHits: counter,
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", handlerFunc)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerNRequests)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerResetRequests)
	serveMux.HandleFunc("POST /api/validate_chirp", handlerChirp)
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverStruct := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err := serverStruct.ListenAndServe()
	fmt.Println(err)
}
