package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"encoding/json"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}

func removeProfane(text string) string {
	splitted := strings.Split(text, " ")
	for i, word := range splitted {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) =="sharbert" || strings.ToLower(word) == "fornax" {
			splitted[i] = "****"
		}	
	}
	return strings.Join(splitted, " ")
}


func SendErrorResponse(w http.ResponseWriter, r *http.Request, text string){
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


func SendValidResponse(w http.ResponseWriter, r *http.Request, text string){
    type returnVals struct {
        cleanedBody string `json:"cleaned_body"`
    }
    respBody := returnVals{
		cleanedBody:removeProfane(text),
    }
    dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
	w.Write(dat)
}


func DecodeHandler(w http.ResponseWriter, r *http.Request){
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
		SendValidResponse(w, r, params.Body)
	}
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

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /api/validate_chirp", DecodeHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}