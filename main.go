package main

import (
	"log"
	"net/http"
	"os"
	"web-portfolio-api/internal"
	"web-portfolio-api/internal/db"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /views", viewsHandler.GetViews)
	mux.HandleFunc("POST /views/add", viewsHandler.IncrementViews)
	mux.HandleFunc("GET /matrix",)

	log.Printf("listening on: %s", port)
	log.Fatal(http.ListenAndServe(":" + port, mux))
}
