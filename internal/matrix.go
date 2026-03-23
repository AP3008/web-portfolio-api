package internal

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"web-portfolio-api/internal/db"
)

type MatrixHandler struct {
	store * db.Store
	apiKey string
}

func NewMatrixHandler (store *db.Store, apiKey string) (*MatrixHandler) {
	return &MatrixHandler {
		store: store, 
		apiKey: apiKey,
	}
}


