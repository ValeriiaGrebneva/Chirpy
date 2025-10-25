package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}

func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/healthz", handlerFunc)
	serveMux.Handle("/app/",  http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serverStruct := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err := serverStruct.ListenAndServe()
	fmt.Println(err)
}
