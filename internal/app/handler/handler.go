package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Skramouch/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	storage *storage.URLStorage
	baseURL string
}

func New(storage *storage.URLStorage, baseURL string) *URLHandler {
	return &URLHandler{
		storage: storage,
		baseURL: baseURL,
	}
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	ShortURL string `json:"short_url"`
}

// CreateShortURL обрабатывает POST-запрос на создание короткого URL
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	id, err := h.storage.Save(req.URL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := createResponse{
		ShortURL: h.baseURL + "/" + id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	url, err := h.storage.Get(id)
	if err == storage.ErrURLNotFound {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}