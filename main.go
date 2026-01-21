package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ashutosh-pradhan777/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const DEVENV = "dev"

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) numRequests(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, `<html> 
						<body>
							<h1>Welcome, Chirpy Admin</h1>
							<p>Chirpy has been visited %d times!</p>
						</body>
						</html>`,
			cfg.fileserverHits.Load())
		w.Header().Set("Content-Type", "text/html")
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetRequests(w http.ResponseWriter, r *http.Request)  {

	if strings.ToLower(cfg.platform) != DEVENV {
		log.Print("403 Forbidden")
		w.WriteHeader(403)
		return
	}
	
	err := cfg.dbQueries.ResetUser(r.Context())
	if err != nil {
		log.Print("Reset Error")
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("users table has been reset."))
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	body,err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	defer r.Body.Close()

	var data map[string]any

	err2 := json.Unmarshal(body,&data)
	if err2 != nil {
		log.Printf("Error decoding parameters: %s", err2)
		w.WriteHeader(500)
		return
	}

	email,ok := data["email"].(string) // type assertion...
	if !ok {
		log.Printf("Can't fetch param. Mention correct keyword for email. ")
		w.WriteHeader(500)
		return
	}

	userData,err := cfg.dbQueries.CreateUser(r.Context(), email)
	if err != nil {
		log.Printf("Error fetching user.")
		w.WriteHeader(500)
		return
	}

	respData := User {
		ID: userData.ID,
		CreatedAt: userData.CreatedAt,
		UpdatedAt: userData.UpdatedAt,
		Email: userData.Email,
	}

	jsonresp,err := json.Marshal(respData)
	if err != nil {
		log.Print("Error handling json marshalling.")
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(201)
	w.Write([]byte(jsonresp))

}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	dbQueries := database.New(db)
	

	mux := http.NewServeMux()

	var cfg apiConfig
	cfg.dbQueries = dbQueries
	cfg.platform = os.Getenv("PLATFORM")

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.Handle("GET /admin/metrics", cfg.numRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	mux.HandleFunc("POST /admin/reset", cfg.resetRequests)

	mux.HandleFunc("POST /api/users", cfg.createUser)

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {

		type reqbody struct {
			Body string `json:"body"`
		}

		type resbody struct {
			Error string `json:"error"`
		}

		type Set map[string]struct{}

		decoder := json.NewDecoder(r.Body)
		var body reqbody

		err := decoder.Decode(&body)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}

		if len(body.Body) > 140 {

			resp := resbody{
				Error: "Chirp is too long",
			}

			data, err := json.Marshal(resp)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(500)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write(data)

			return
		}

		badwords := []string{"kerfuffle", "sharbert", "fornax"}
		bwords := make(Set)
		for _, ele := range badwords {
			bwords[ele] = struct{}{}
		}

		var finalres string
		var finalarr []string
		allwords := strings.Split(body.Body, " ")
		for _, word := range allwords {

			nword := strings.ToLower(word)
			if _, ok := bwords[nword]; ok {

				finalarr = append(finalarr, "****")
				continue

			}
			finalarr = append(finalarr, word)

		}

		finalres = strings.Join(finalarr, " ")

		type validresp struct {
			CleanedBody string `json:"cleaned_body"`
		}

		resp := validresp{
			CleanedBody: finalres,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		w.Write(data)
	})

	server := &http.Server{

		Handler: mux,
		Addr:    ":8080",
	}

	errx := server.ListenAndServe()
	if errx != nil {
		fmt.Printf("Error: %v\n", errx)
	}

}

// go build -o out && ./out
