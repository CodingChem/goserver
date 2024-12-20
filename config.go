package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/codingchem/goserver/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(rw, req)
	})
}

func (cfg *apiConfig) getFileServerHits(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(200)
	hits := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	rw.Write([]byte(hits))
}

func (cfg *apiConfig) resetServerHits(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}

func NewApiConfig(db *database.Queries) apiConfig {
	return apiConfig{
		db: db,
	}
}
