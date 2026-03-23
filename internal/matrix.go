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

type matrixRepsonse struct {
	Rows  int     `json:"rows"`
	Cols  int     `json:"cols"`
	Cells [][]int `json:"cells"`
}

type toggleRequest struct {
	Row  int     `json:"row"`
	Col  int     `json:"col"`
}

type toggleResponse struct {
	Row   int     `json:"row"`
	Col   int     `json:"col"`
	Value int     `json:"value"`
}

func (h *MatrixHandler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	cells, err := h.store.GetMatrix(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, 
		http.StatusInternalServerError)
	}
	w.Header().Set("Context-Type", "application/json")
	json.NewEncoder(w).Encode(matrixRepsonse{
		Rows: db.MatrixSize,
		Cols: db.MatrixSize,
		Cells: cells,
	})
}

func (h *MatrixHandler) ToggleCell(w http.ResponseWriter, r *http.Request){
	provided := r.URL.Query().Get("key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.apiKey)) != 1{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error":"unauthorized"})
		return
	}
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return 
	}
	if req.Row < 0 || req.Row >= db.MatrixSize || req.Col < 0 || req.Col >= db.MatrixSize {
		http.Error(w, `{"error":"row or col out of bounds"}`, http.StatusBadRequest)
		return 
	}

	value, err := h.store.ToggleCell(r.Context(), req.Row, req.Col)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return 
	}

	w.Header().Set("Content-Type", "application/json"
	json.NewEncoder(w).Encode(toggleResponse{
		Row:   req.Row, 
		Col:   req.Col,
		Value: value, 
	})
}
