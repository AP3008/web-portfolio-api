package main

import (
	"log"
	"net/http"
	"os"
	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"

	"github.com/go-chi/chi"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	port := "8080"
	dbPath := "views.db"
	
	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	viewsHandler := internal.ViewsHandler(store, apiKey)
	r := chi.NewRouter()

	r.Use
}
