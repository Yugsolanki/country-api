package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Yugsolanki/country-api/internal/models"
	"github.com/Yugsolanki/country-api/internal/service"
)

// handles http requests for the country operations
type CountryHandler struct {
	service *service.CountryService
	logger  *log.Logger
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// creates a new country handler
func NewCountryHandler(svc *service.CountryService, logger *log.Logger) *CountryHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &CountryHandler{
		service: svc,
		logger:  logger,
	}
}

// handles GET /api/countries/search
func (h *CountryHandler) Search(w http.ResponseWriter, r *http.Request) {
	// only allow GET requests
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET method is supported")
		return
	}

	// get the name parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid request", "Query parameter 'name' is required")
		return
	}

	h.logger.Printf("Searching for country: %s", name)

	// search for the country
	country, err := h.service.SearchCountry(r.Context(), name)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, country)
}

func (h *CountryHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrCountryNotFound):
		h.writeError(w, http.StatusNotFound, "Not found", "Country not found")
	case errors.Is(err, models.ErrInvalidRequest):
		h.writeError(w, http.StatusBadRequest, "Invalid request", "Invalid request")
	case errors.Is(err, models.ErrTimeout):
		h.writeError(w, http.StatusGatewayTimeout, "Timeout", "Request to external service timed out")
	case errors.Is(err, models.ErrAPIFailure):
		h.writeError(w, http.StatusBadGateway, "Service unavailable", "External service is unavailable")
	default:
		h.logger.Printf("Unexpected error: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Internal error", "An unexpected error occurred")
	}
}

// writes a json response
func (h *CountryHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
	}
}

// writes an error response
func (h *CountryHandler) writeError(w http.ResponseWriter, status int, error string, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error:   error,
		Message: message,
	})
}
