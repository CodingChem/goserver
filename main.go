package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle(
		"/",
		http.FileServer(
				http.Dir("."),
		),
	)
	mux.HandleFunc("/healthz", healtWriter)
	server := http.Server {Addr: ":8080",Handler: mux}
	server.ListenAndServe()
}

func healtWriter(w http.ResponseWriter,req *http.Request) {
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
