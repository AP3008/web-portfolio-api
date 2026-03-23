package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"
)

var (
	mux  *http.ServeMux
	once sync.Once
)

func setup() {
	tursoURL := os.Getenv("TURSO_DB_URL")
	tursoToken := os.Getenv("TURSO_DB_AUTH_TOKEN")
	apiKey := os.Getenv("API_KEY")

	if tursoURL == "" || tursoToken == "" {
		log.Fatal("TURSO_DB_URL and TURSO_DB_AUTH_TOKEN must be set")
	}
	if apiKey == "" {
		log.Fatal("API_KEY must be set")
	}

	store, err := db.OpenTurso(tursoURL, tursoToken)
	if err != nil {
		log.Fatalf("Failed to open Turso database: %v", err)
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
