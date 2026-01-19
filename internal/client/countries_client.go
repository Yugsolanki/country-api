package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Yugsolanki/country-api/internal/models"
)

type CountriesClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
	Logger  *log.Logger
}

func NewCountriesClient(config ClientConfig) *CountriesClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://restcountries.com/v3.1"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Logger == nil {
		config.Logger = log.Default()
	}

	return &CountriesClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: config.Logger,
	}
}

func (c *CountriesClient) SearchByName(ctx context.Context, name string) (*models.Country, error) {
	if name == "" {
		return nil, models.ErrInvalidRequest
	}

	// Building the url
	// example: https://restcountries.com/v3.1/name/India?fullText=true
	endpoint := fmt.Sprintf("%s/name/%s?fullText=true", c.baseURL, name)

	c.logger.Printf("Fetching country date from: %s", endpoint)

	// creat request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		c.logger.Printf("Error creating request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.logger.Printf("Request timeout for country: %s", name)
			return nil, models.ErrTimeout
		}
		c.logger.Printf("Error executing request: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrAPIFailure, err)
	}
	defer resp.Body.Close()

	// handle response status
	if resp.StatusCode == http.StatusNotFound {
		c.logger.Printf("Country not found: %s", name)
		return nil, models.ErrCountryNotFound
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Printf("API returned status %d for country: %s", resp.StatusCode, name)
		return nil, fmt.Errorf("%w: status code %d", models.ErrAPIFailure, resp.StatusCode)
	}

	// read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Printf("Error reading response from body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var countries []models.RestCountryResponse
	if err := json.Unmarshal(body, &countries); err != nil {
		c.logger.Printf("Error parsing response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(countries) == 0 {
		return nil, models.ErrCountryNotFound
	}

	// convert to our model
	return c.convertToCountry(&countries[0]), nil
}

func (c *CountriesClient) convertToCountry(resp *models.RestCountryResponse) *models.Country {
	country := &models.Country{
		Name:       resp.Name.Common,
		Population: resp.Population,
	}

	// get capital (taking first if multiple)
	if len(resp.Capital) > 0 {
		country.Capital = resp.Capital[0]
	}

	// get currency symbol (taking first if multiple)
	for _, currency := range resp.Currencies {
		country.Currency = currency.Symbol
		break
	}

	return country
}
