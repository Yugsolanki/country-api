package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Yugsolanki/country-api/internal/models"
)

func TestCountriesClient_SearchByBame_Success(t *testing.T) {
	// create mock server
	mockResponse := []models.RestCountryResponse{
		{
			Name: models.NameInfo{
				Common:   "India",
				Official: "Republic of India",
			},
			Capital:    []string{"New Delhi"},
			Population: 1417492000,
			Currencies: map[string]models.Currency{
				"INR": {Name: "Indian rupee", Symbol: "₹"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// create client
	client := NewCountriesClient(ClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		Logger:  log.New(os.Stdout, "", 0),
	})

	// test
	country, err := client.SearchByName(context.Background(), "India")
	if err != nil {
		t.Fatalf("Expected no error got %v", err)
	}

	if country.Name != "India" {
		t.Errorf("Expected name 'India' got '%s'", country.Name)
	}
	if country.Capital != "New Delhi" {
		t.Errorf("Expected capital 'New Delhi' got '%s'", country.Capital)
	}
	if country.Currency != "₹" {
		t.Errorf("Expected currency '₹' got %s", country.Currency)
	}
	if country.Population != 1417492000 {
		t.Errorf("Expected population 1417492000 got %d", country.Population)
	}
}

func TestCountriesClient_SearchByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewCountriesClient(ClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		Logger:  log.New(os.Stdout, "", 0),
	})

	_, err := client.SearchByName(context.Background(), "Nonexistent")
	if err != models.ErrCountryNotFound {
		t.Errorf("Expected ErrCountryNotFound go %v", err)
	}
}

func TestCountriesClient_SearchByName_EmptyName(t *testing.T) {
	client := NewCountriesClient(ClientConfig{
		Timeout: 5 * time.Second,
		Logger:  log.New(os.Stdout, "", 0),
	})

	_, err := client.SearchByName(context.Background(), "")
	if err != models.ErrInvalidRequest {
		t.Errorf("Expected ErrInvalidRequest got %v", err)
	}
}

func TestCountriesClient_SearchByName_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	client := NewCountriesClient(ClientConfig{
		BaseURL: server.URL,
		Timeout: 100 * time.Millisecond,
		Logger:  log.New(os.Stdout, "", 0),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.SearchByName(ctx, "India")
	if err == nil {
		t.Error("Expected timeout err")
	}
}

func TestCountriesClient_SearchByName_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewCountriesClient(ClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		Logger:  log.New(os.Stdout, "", 0),
	})

	_, err := client.SearchByName(context.Background(), "India")
	if err == nil {
		t.Error("Expected error for server error response")
	}
}
