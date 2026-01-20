package service

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/Yugsolanki/country-api/internal/models"
)

// Mockcache implementation for testing
type MockCache struct {
	data map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *MockCache) Set(key string, value interface{}) {
	m.data[key] = value
}

// implements a mock for CountriesClient
type MockClient struct {
	SearchFunc func(ctx context.Context, name string) (*models.Country, error)
}

func (m *MockClient) SearchByName(ctx context.Context, name string) (*models.Country, error) {
	return m.SearchFunc(ctx, name)
}

// interface for testing
type ClientInterface interface {
	SearchByName(ctx context.Context, name string) (*models.Country, error)
}

// wraps CountryService for testing
type TestableCountryService struct {
	client ClientInterface
	cache  *MockCache
	logger *log.Logger
}

func NewTestableService(client ClientInterface, cache *MockCache) *TestableCountryService {
	return &TestableCountryService{
		client: client,
		cache:  cache,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (s *TestableCountryService) SearchCountry(ctx context.Context, name string) (*models.Country, error) {
	if name == "" {
		return nil, models.ErrInvalidRequest
	}

	cacheKey := name

	if cached, found := s.cache.Get(cacheKey); found {
		if country, ok := cached.(*models.Country); ok {
			return country, nil
		}
	}

	country, err := s.client.SearchByName(ctx, name)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, country)
	return country, nil
}

func TestCountryService_SearchCountry_CacheHit(t *testing.T) {
	cache := NewMockCache()
	expectedCountry := &models.Country{
		Name:       "India",
		Capital:    "New Delhi",
		Currency:   "₹",
		Population: 1417492000,
	}
	cache.Set("india", expectedCountry)

	MockClient := &MockClient{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			t.Error("Client should not be called on cache hit")
			return nil, nil
		},
	}

	service := NewTestableService(MockClient, cache)

	country, err := service.SearchCountry(context.Background(), "india")
	if err != nil {
		t.Fatalf("Expected no error but go %v", err)
	}

	if country.Name != expectedCountry.Name {
		t.Errorf("Expected name %s but got %s", expectedCountry.Name, country.Name)
	}
}

func TestCountryService_SearchCountry_CacheMiss(t *testing.T) {
	cache := NewMockCache()
	expectedCountry := &models.Country{
		Name:       "India",
		Capital:    "New Delhi",
		Currency:   "₹",
		Population: 1417492000,
	}

	clientCalled := false
	mockClient := &MockClient{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			clientCalled = true
			return expectedCountry, nil
		},
	}

	service := NewTestableService(mockClient, cache)

	country, err := service.SearchCountry(context.Background(), "India")
	if err != nil {
		t.Fatalf("Expected no error but go %v", err)
	}

	if !clientCalled {
		t.Error("Expected client to be called on cache miss")
	}

	if country.Name != expectedCountry.Name {
		t.Errorf("Expected name %s but go %s", expectedCountry.Name, country.Name)
	}

	// verify it was cached
	if _, found := cache.Get("India"); !found {
		t.Error("Expected country to be cached after fetch")
	}
}

func TestCountryService_SearchCountry_EmptyName(t *testing.T) {
	cache := NewMockCache()
	mockClient := &MockClient{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			return nil, nil
		},
	}

	service := NewTestableService(mockClient, cache)

	_, err := service.SearchCountry(context.Background(), "")
	if !errors.Is(err, models.ErrInvalidRequest) {
		t.Errorf("Expected ErrInvalidRequest but got %v", err)
	}
}

func TestCountryService_SearchCountry_ClientError(t *testing.T) {
	cache := NewMockCache()
	mockClient := &MockClient{
		SearchFunc: func(ctx context.Context, name string) (*models.Country, error) {
			return nil, models.ErrCountryNotFound
		},
	}

	service := NewTestableService(mockClient, cache)

	_, err := service.SearchCountry(context.Background(), "Nonexistent")
	if !errors.Is(err, models.ErrCountryNotFound) {
		t.Errorf("Expected ErrCountryNotFound but got %v", err)
	}
}
