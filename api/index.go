package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"

	_ "github.com/lib/pq"
)

var (
	mux  *http.ServeMux
	once sync.Once
)

func setup() {
	postgresURL := os.Getenv("POSTGRES_URL")
	apiKey := os.Getenv("API_KEY")

	if postgresURL == "" {
		log.Fatal("POSTGRES_URL must be set")
	}
	if apiKey == "" {
		log.Fatal("API_KEY must be set")
	}

	store, err := db.OpenPostgres(postgresURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := store.InitMatrix(context.Background()); err != nil {
		log.Fatalf("Failed to initialize matrix: %v", err)
	}

	viewsHandler := internal.NewViewsHandler(store, apiKey)
	matrixHandler := internal.NewMatrixHandler(store, apiKey)

	mux = http.NewServeMux()
	mux.HandleFunc("GET /views", viewsHandler.GetViews)
	mux.HandleFunc("POST /views/add", viewsHandler.IncrementViews)
	mux.HandleFunc("GET /matrix", matrixHandler.GetMatrix)
	mux.HandleFunc("POST /matrix/toggle", matrixHandler.ToggleCell)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(setup)
	mux.ServeHTTP(w, r)
}
