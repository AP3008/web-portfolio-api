package main

import (
	"log"
	"net/http"
	"os"
	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	port := "8080"
	dbPath := "views.db"
	
	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	viewsHandler := internal.NewViewsHandler(store, apiKey)
	r := chi.NewRouter()
	r.Get("/views", viewsHandler.GetViews)
	r.Post("/views", viewsHandler.IncrementViews)
	log.Printf("listening on: %s", port)
	log.Fatal(http.ListenAndServe(":" + port, r))
}
