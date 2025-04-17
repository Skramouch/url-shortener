package storage

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
)

var (
	ErrURLNotFound = errors.New("url not found")
)

// URLStorage представляет хранилище URL
type URLStorage struct {
	mu       sync.RWMutex
	urls     map[string]string // id -> original URL
}

// New создает новое хранилище URL
func New() *URLStorage {
	return &URLStorage{
		urls: make(map[string]string),
	}
}

// Save сохраняет оригинальный URL и возвращает сгенерированный ID
func (s *URLStorage) Save(originalURL string) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.urls[id] = originalURL
	s.mu.Unlock()

	return id, nil
}

// Get возвращает оригинальный URL по ID
func (s *URLStorage) Get(id string) (string, error) {
	s.mu.RLock()
	url, exists := s.urls[id]
	s.mu.RUnlock()

	if !exists {
		return "", ErrURLNotFound
	}

	return url, nil
}

// generateID генерирует уникальный идентификатор для URL
func generateID() (string, error) {
	b := make([]byte, 6) // 6 байт дадут 8 символов в base64
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}