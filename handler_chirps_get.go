package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/Groskilled/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, database.Chirp{
		ID:       dbChirp.ID,
		Body:     dbChirp.Body,
		AuthorId: dbChirp.AuthorId,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	authId := -1
	var err error
	stringID := r.URL.Query().Get("author_id")
	if stringID != "" {
		authId, err = strconv.Atoi(stringID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't retrieve chirps")
			return
		}
	}
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []database.Chirp{}
	for _, dbChirp := range dbChirps {
		if authId != -1 && dbChirp.AuthorId != authId {
			continue
		}
		chirps = append(chirps, database.Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}
