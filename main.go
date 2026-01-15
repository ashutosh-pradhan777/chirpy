package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) numRequests(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetRequests(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Store(0)
		next.ServeHTTP(w, r)
	})
}

func main() {

	mux := http.NewServeMux()

	var cfg apiConfig

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.Handle("/metrics", cfg.numRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	mux.Handle("/reset", cfg.resetRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	server := &http.Server{

		Handler: mux,
		Addr:    ":8080",
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

}

// go build -o out && ./out
