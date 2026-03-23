package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"

	_ "modernc.org/sqlite"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}
	port := "8080"
	dbPath := "views.db"
	
	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := store.InitMatrix(context.Background()); err != nil {
		log.Fatalf("Failed to initialize matrix: %v", err)
	}

	viewsHandler := internal.NewViewsHandler(store, apiKey)
	
	matrixHandler := internal.NewMatrixHandler(store, apiKey)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /views", viewsHandler.GetViews)
	mux.HandleFunc("POST /views/add", viewsHandler.IncrementViews)
	mux.HandleFunc("GET /matrix", matrixHandler.GetMatrix)
	mux.HandleFunc("POST /matrix/toggle", matrixHandler.ToggleCell)

	log.Printf("listening on: %s", port)
	log.Fatal(http.ListenAndServe(":" + port, mux))
}
