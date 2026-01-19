package service

import (
	"context"
	"log"
	"strings"

	"github.com/Yugsolanki/country-api/internal/cache"
	"github.com/Yugsolanki/country-api/internal/client"
	"github.com/Yugsolanki/country-api/internal/models"
)

// Service to handle buisness logic
type CountryService struct {
	client *client.CountriesClient
	cache  cache.Cache
	logger *log.Logger
}

// holds configuration for the service
type ServiceConfig struct {
	Client *client.CountriesClient
	Cache  cache.Cache
	Logger *log.Logger
}

// creates a new country service instance
func NewCountryService(config ServiceConfig) *CountryService {
	if config.Logger == nil {
		config.Logger = log.Default()
	}

	return &CountryService{
		client: config.Client,
		cache:  config.Cache,
		logger: config.Logger,
	}
}

// searches for a country by name, using cache when available
func (s *CountryService) SearchCountry(ctx context.Context, name string) (*models.Country, error) {
	// validate input
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, models.ErrInvalidRequest
	}

	// normalize the cache key
	cacheKey := strings.ToLower(name)

	// try to get from cache first
	if cached, found := s.cache.Get(cacheKey); found {
		s.logger.Printf("Cache hit for country: %s", name)
		if country, ok := cached.(*models.Country); ok {
			return country, nil
		}
	}

	// fetch from external api
	country, err := s.client.SearchByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// store in cache
	s.cache.Set(cacheKey, country)
	s.logger.Printf("Cached country data for: %s", name)

	return country, nil
}
