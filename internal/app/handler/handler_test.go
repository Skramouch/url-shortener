package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Skramouch/url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// Мок хранилища URL
type mockURLStorage struct {
	saveFn func(string) (string, error)
	getFn  func(string) (string, error)
}

func (m *mockURLStorage) Save(url string) (string, error) {
	return m.saveFn(url)
}

func (m *mockURLStorage) Get(id string) (string, error) {
	return m.getFn(id)
}

// Тестовый вариант URLHandler с интерфейсом вместо конкретной реализации хранилища
type testURLHandler struct {
	URLHandler
	storageInterface interface {
		Save(string) (string, error)
		Get(string) (string, error)
	}
}

// Переопределяем методы для использования storageInterface
func (h *testURLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.storageInterface.Save(url)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	shortURL := h.baseURL + "/" + id

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *testURLHandler) GetOriginalURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	url, err := h.storageInterface.Get(id)
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

func TestCreateShortURL(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSaveFn     func(string) (string, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Успешное создание короткого URL",
			requestBody: "https://example.com",
			mockSaveFn: func(url string) (string, error) {
				return "abc123", nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc123",
		},
		{
			name:        "Пустой URL",
			requestBody: "",
			mockSaveFn: func(url string) (string, error) {
				return "", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "URL обязателен\n",
		},
		{
			name:        "Ошибка при сохранении",
			requestBody: "https://example.com",
			mockSaveFn: func(url string) (string, error) {
				return "", errors.New("ошибка сохранения")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Внутренняя ошибка сервера\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем мок хранилища
			mockStorage := &mockURLStorage{
				saveFn: tc.mockSaveFn,
			}

			// Создаем тестовый обработчик
			h := &testURLHandler{
				URLHandler: URLHandler{
					baseURL: "http://localhost:8080",
				},
				storageInterface: mockStorage,
			}

			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.requestBody))
			rec := httptest.NewRecorder()

			// Выполняем запрос
			h.CreateShortURL(rec, req)

			// Проверяем результаты
			if rec.Code != tc.expectedStatus {
				t.Errorf("Ожидается код статуса %d, получен %d", tc.expectedStatus, rec.Code)
			}

			if rec.Body.String() != tc.expectedBody {
				t.Errorf("Ожидается тело ответа %q, получено %q", tc.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	tests := []struct {
		name           string
		urlID          string
		mockGetFn      func(string) (string, error)
		expectedStatus int
		expectedURL    string
	}{
		{
			name:  "Успешное получение оригинального URL",
			urlID: "abc123",
			mockGetFn: func(id string) (string, error) {
				return "https://example.com", nil
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    "https://example.com",
		},
		{
			name:  "URL не найден",
			urlID: "notfound",
			mockGetFn: func(id string) (string, error) {
				return "", storage.ErrURLNotFound
			},
			expectedStatus: http.StatusBadRequest,
			expectedURL:    "",
		},
		{
			name:  "Внутренняя ошибка",
			urlID: "abc123",
			mockGetFn: func(id string) (string, error) {
				return "", errors.New("внутренняя ошибка")
			},
			expectedStatus: http.StatusBadRequest,
			expectedURL:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем мок хранилища
			mockStorage := &mockURLStorage{
				getFn: tc.mockGetFn,
			}

			// Создаем тестовый обработчик
			h := &testURLHandler{
				URLHandler: URLHandler{
					baseURL: "http://localhost:8080",
				},
				storageInterface: mockStorage,
			}

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/"+tc.urlID, nil)

			// Настраиваем chi router для работы с URLParam
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tc.urlID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			rec := httptest.NewRecorder()

			// Выполняем запрос
			h.GetOriginalURL(rec, req)

			// Проверяем результаты
			if rec.Code != tc.expectedStatus {
				t.Errorf("Ожидается код статуса %d, получен %d", tc.expectedStatus, rec.Code)
			}

			if tc.expectedURL != "" && rec.Header().Get("Location") != tc.expectedURL {
				t.Errorf("Ожидается URL %q, получен %q", tc.expectedURL, rec.Header().Get("Location"))
			}
		})
	}
}