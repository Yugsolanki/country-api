package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Yugsolanki/country-api/internal/models"
)

// mocks the service layer
type MockCountryService struct {
	SearchFunc func(ctx context.Context, name string) (*models.Country, error)
}

func (m *MockCountryService) SearchCountry(ctx context.Context, name string) (*models.Country, error) {
	return m.SearchFunc(ctx, name)
}

// we can inject mock into handler
type ServiceInterface interface {
	SearchCountry(ctx context.Context, name string) (*models.Country, error)
}

// for testing with mock service
type TestableCountryHandler struct {
	service ServiceInterface
	logger  *log.Logger
}

func NewTestableHandler(service ServiceInterface) *TestableCountryHandler {
	return &TestableCountryHandler{
		service: service,
		logger:  log.New(os.Stdout, "", 0),
	}
}

func (h *TestableCountryHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("Failed to encode response: %v", err)
	}
}

func (h *TestableCountryHandler) writeError(w http.ResponseWriter, status int, error string, message string) {
	h.writeJSON(w, status, ErrorResponse{Error: error, Message: message})
}

func (h *TestableCountryHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET method is allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid request", "Name is required")
		return
	}

	country, err := h.service.SearchCountry(r.Context(), name)
	if err != nil {
		switch err {
		case models.ErrCountryNotFound:
			h.writeError(w, http.StatusNotFound, "Not found", "Country not foubd")
		case models.ErrInvalidRequest:
			h.writeError(w, http.StatusBadRequest, "Invalid request", err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, country)
}

func TestCountryHandler_Search_Success(t *testing.T) {
	expectedCountry := &models.Country{
		Name:       "India",
		Capital:    "New Delhi",
		Currency:   "₹",
		Population: 1417492000,
	}

	mockService := &MockCountryService{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			return expectedCountry, nil
		},
	}

	handler := NewTestableHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/countries/search?name=India", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 but got %d", rec.Code)
	}

	var country models.Country
	if err := json.NewDecoder(rec.Body).Decode(&country); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if country.Name != expectedCountry.Name {
		t.Errorf("Expected name %s but go %s", expectedCountry.Name, country.Name)
	}
}

func TestCountryHandler_Search_MissingName(t *testing.T) {
	mockService := &MockCountryService{}
	handler := NewTestableHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/countries/search", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 but got %d", rec.Code)
	}
}

func TestCountryHandler_Search_NotFound(t *testing.T) {
	mockService := &MockCountryService{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			return nil, models.ErrCountryNotFound
		},
	}

	handler := NewTestableHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/countries/search?name=Nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 but go %d", rec.Code)
	}
}

func TestCountryHandler_Search_WrongMethod(t *testing.T) {
	mockService := &MockCountryService{}
	handler := NewTestableHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/countries/search?name=India", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 but got %d", rec.Code)
	}
}

func TestCountryHandler_Search_ContentType(t *testing.T) {
	mockService := &MockCountryService{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			return &models.Country{Name: "India"}, nil
		},
	}

	handler := NewTestableHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/countries/search?name=India", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
