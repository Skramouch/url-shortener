package handler

import (
	"io"
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

// CreateShortURL обрабатывает POST-запрос на создание короткого URL
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	url := string(body)
	if url == "" {
		http.Error(w, "URL обязателен", http.StatusBadRequest)
		return
	}

	id, err := h.storage.Save(url)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	shortURL := h.baseURL + "/" + id

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
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