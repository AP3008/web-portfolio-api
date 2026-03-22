package internal

import (
	"crypto/subtle"
	_ "crypto/subtle"
	"encoding/json"
	_ "encoding/json"
	"net/http"
	"web-portfolio-api/internal/db"
)

type ViewsHandler struct {
	store * db.Store
	apiKey string
}

func NewViewsHandle(store *db.Store, apiKey string) *ViewsHandler {
	return &ViewsHandler{
		store: store,
		apiKey: apiKey,
	}
}

type viewsResponse struct {
	Count int64 `json: "count"`
}

func (h *ViewsHandler) GetViews(w http.ResponseWriter, r *http.Request){
	count, err := h.store.GetCount(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"internal error"}`,
			http.StatusInternalServerError,
		)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(viewsResponse{Count: count})
}

func (h *ViewsHandler) IncrementViews(w http.ResponseWriter, r *http.Request){
	provided := r.URL.Query().Get("key")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.key)){
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnaithorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return 
	}
	count, err := h.store.Increment(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(viewsResponse{Count: count})
}
