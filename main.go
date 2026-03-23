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

	viewsHandler := handler.viewsHandler
}
