package main

import (
	"fmt"
	"net/http"
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
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverStruct := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err := serverStruct.ListenAndServe()
	fmt.Println(err)
}
