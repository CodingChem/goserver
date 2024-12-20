package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
)

type responseError struct {
	Error string `json:"error"`
}

func main() {
	mux := http.NewServeMux()
	conf := apiConfig{fileserverHits: atomic.Int32{}}
	mux.Handle(
		"/app/",
		conf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(
			http.Dir("."),
		))),
	)
	mux.HandleFunc("GET /api/healthz", healtWriter)
	mux.HandleFunc("POST /api/validate_chirp", chirpValidator)
	mux.HandleFunc("GET /admin/metrics", conf.getFileServerHits)
	mux.HandleFunc("POST /admin/reset", conf.resetServerHits)
	server := http.Server{Addr: ":8080", Handler: mux}
	fmt.Printf("Starting server...")
	server.ListenAndServe()
}

func healtWriter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func chirpValidator(rw http.ResponseWriter, req *http.Request) {
	type requestStruct struct {
		Body string `json:"body"`
	}
	type responseClean struct {
		CleanBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	params := requestStruct{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		// Chirp is too long
		respondWithError(rw, 400, "Chirp is too long")
		return
	} else {
		// Chirp is not to long
		respondWithJSON(rw, 200, responseClean{CleanBody: replaceBadWords(params.Body)})
		return
	}
}
func respondWithError(rw http.ResponseWriter, code int, msg string) {
	responseStruct := responseError{Error: msg}
	responseData, err := json.Marshal(responseStruct)
	if err != nil {
		log.Printf("Error encoding static error response in respondWithError: %s", err)
		rw.WriteHeader(500)
		return
	}
	rw.WriteHeader(code)
	rw.Write(responseData)
	return
}
func respondWithJSON(rw http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding payload in respondWithJSON: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	rw.WriteHeader(code)
	rw.Write(data)
	return
}

func replaceBadWords(msg string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	msgSlice := strings.Split(msg, " ")
	lowerSlice := strings.Split(strings.ToLower(msg), " ")
	returnSlice := make([]string, len(msgSlice))
	for idx, value := range lowerSlice {
		if !slices.Contains(badWords, value) {
			returnSlice = append(returnSlice, msgSlice[idx])
		} else {
			returnSlice = append(returnSlice, "****")
		}
	}
	return strings.Trim(strings.Join(returnSlice, " "), " ")
}
