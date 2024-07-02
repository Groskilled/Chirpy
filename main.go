package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Groskilled/Chirpy/internal/chirps"
	"github.com/Groskilled/Chirpy/internal/database"
	"github.com/Groskilled/Chirpy/internal/users"

	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits int
	index          int
	mutex          sync.Mutex
	db             database.DB
}

func removeProfane(text string) string {
	splitted := strings.Split(text, " ")
	for i, word := range splitted {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			splitted[i] = "****"
		}
	}
	res := strings.Join(splitted, " ")
	return res
}

func SendErrorResponse(w http.ResponseWriter, r *http.Request, text string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: text,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	w.Write(dat)
}

func (cfg *apiConfig) SendValidResponse(w http.ResponseWriter, r *http.Request, text string) {
	type returnVals struct {
		Id          int    `json:"id"`
		CleanedBody string `json:"body"`
	}

	cfg.mutex.Lock()
	cfg.index++
	id := cfg.index
	cfg.mutex.Unlock()

	respBody := returnVals{
		Id:          id,
		CleanedBody: removeProfane(text),
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	_, err = chirps.CreateChirp(cfg.db, respBody.Id, respBody.CleanedBody)
	if err != nil {
		fmt.Printf("Error while creating Chirp: %s\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := chirps.GetChirps(cfg.db)
	if err != nil {
		fmt.Printf("Error while loading DB: %s\n", err)
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	dat, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, er := strconv.Atoi(r.PathValue("chirpID"))
	if er != nil {
		fmt.Printf("Atoi Failed: %s\n", er)
		w.WriteHeader(500)
		return
	}
	chirp, err := chirps.GetChirpById(cfg.db, chirpId)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	dat, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
	return
}

func (cfg *apiConfig) SendValidUser(w http.ResponseWriter, r *http.Request, email string, pw string) {
	type returnVals struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
		//Password string `json:"password"`
	}

	cfg.mutex.Lock()
	cfg.index++
	id := cfg.index
	cfg.mutex.Unlock()

	respBody := returnVals{
		Id:    id,
		Email: email,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(pw), 14)
	_, err = users.CreateUser(cfg.db, respBody.Id, respBody.Email, string(hashedPw))
	if err != nil {
		fmt.Printf("Error while creating User: %s\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) PostUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		SendErrorResponse(w, r, "Something went wrong")
		return
	}
}

func (cfg *apiConfig) DecodeHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		SendErrorResponse(w, r, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		SendErrorResponse(w, r, "Chirp is too long")
	} else {
		cfg.SendValidResponse(w, r, params.Body)
	}
}

func (cfg *apiConfig) DecodeUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		SendErrorResponse(w, r, "Something went wrong")
		return
	}
	cfg.SendValidUser(w, r, params.Email, params.Password)
}

func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		SendErrorResponse(w, r, "Something went wrong")
		return
	}
	usr, _ := users.GetUserByEmail(cfg.db, params.Email)
	if usr.Email == "" {
		SendErrorResponse(w, r, "Unknown User")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(params.Password))
	if err != nil {
		w.WriteHeader(401)
		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Unauthorized",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}
	type UserResponse struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}
	response := UserResponse{
		Id:    usr.Id,
		Email: usr.Email,
	}
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := os.ReadFile("metrics.html")
	if err != nil {
		log.Printf("Error reading template file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formattedHTML := fmt.Sprintf(string(htmlContent), cfg.fileserverHits)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(formattedHTML))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		fmt.Printf("Couldnt create DB. Error: %s \n", err)
		return
	}
	apiCfg := apiConfig{
		fileserverHits: 0,
		index:          0,
		db:             *db,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /api/chirps", apiCfg.DecodeHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)
	mux.HandleFunc("POST /api/users", apiCfg.DecodeUser)
	mux.HandleFunc("POST /api/login", apiCfg.Login)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
