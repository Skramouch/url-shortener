package main

import (
	"net/http"

	"github.com/Skramouch/url-shortener/internal/app/handler"
	"github.com/Skramouch/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const baseURL = "http://localhost:8080"

func run(r chi.Router) error {
	return http.ListenAndServe(`:8080`, r)
}

func main() {
	// Инициализируем хранилище и обработчик
	store := storage.New()
	h := handler.New(store, baseURL)

	// Создаем роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetOriginalURL)

	if err := run(r); err != nil {
		panic(err)
	}
}

