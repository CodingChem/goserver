package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	conf := apiConfig{fileserverHits: atomic.Int32{}}
	mux.Handle(
		"/app/",
		conf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(
			http.Dir("."),
		))),
	)
	mux.HandleFunc("GET /healthz", healtWriter)
	mux.HandleFunc("GET /metrics", conf.getFileServerHits)
	mux.HandleFunc("POST /reset", conf.resetServerHits)
	server := http.Server{Addr: ":8080", Handler: mux}
	fmt.Printf("Starting server...")
	server.ListenAndServe()
}

func healtWriter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
