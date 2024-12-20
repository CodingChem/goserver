package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/codingchem/goserver/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	platform       string
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

func (cfg *apiConfig) resetServerHits(rw http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(rw, 403, "Forbidden")
		return
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	cfg.db.DeleteUsers(req.Context())
	return
}

func NewApiConfig(db *database.Queries, platform string) apiConfig {
	return apiConfig{
		db:       db,
		platform: platform,
	}
}

func (cfg *apiConfig) handleCreateUser(rw http.ResponseWriter, req *http.Request) {
	type reqStruct struct {
		Email string `json:"email"`
	}
	var reqData reqStruct
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&reqData); err != nil {
		log.Printf("Error parsing requers: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	user, err := cfg.db.CreateUser(req.Context(), reqData.Email)
	if err != nil {
		log.Printf("Error executing sql: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	type responseUser struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resUser := responseUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(rw, 201, resUser)
}

func (cfg *apiConfig) handleCreateChirp(rw http.ResponseWriter, req *http.Request) {
	type paramStruct struct {
		Chirp  string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	var params paramStruct
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding params in handleCreateChirp: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Chirp,
		UserID: params.UserId,
	})
	if err != nil {
		log.Printf("Error saving to db: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}
	type responsStruct struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	response := responsStruct{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(rw, 201, response)
	return
}

func (cfg *apiConfig) handleGetChirps(rw http.ResponseWriter, req *http.Request) {
	//TODO:
	chirps, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		log.Printf("Error getting chirps from db: %s", err)
		respondWithError(rw, 500, "Something went wrong")
	}
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	var res []Chirp
	for _, el := range chirps {
		res = append(res, Chirp{ID: el.ID, CreatedAt: el.CreatedAt, UpdatedAt: el.UpdatedAt, Body: el.Body, UserID: el.UserID})
	}
	respondWithJSON(rw, 200, res)
	return
}
