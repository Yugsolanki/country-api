package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Yugsolanki/country-api/internal/cache"
	"github.com/Yugsolanki/country-api/internal/client"
	"github.com/Yugsolanki/country-api/internal/handler"
	"github.com/Yugsolanki/country-api/internal/service"
)

func main() {
	// Logger
	logger := log.New(os.Stdout, "[country-api] ", log.LstdFlags|log.Lshortfile)

	// Configuration
	serverAddr := ":8000"
	cacheTTL := 5 * time.Minute
	clientTimeout := 10 * time.Second

	logger.Printf("Starting server with configuration:")

	logger.Printf("Server Address: %s", serverAddr)
	logger.Printf("Cache TTL: %s", cacheTTL)
	logger.Printf("Client Timeout: %s", clientTimeout)

	// Initialize cache
	inMemoryCache := cache.NewInMemoryCache(cacheTTL)
	defer inMemoryCache.Stop()

	// HTTP client
	countriesClient := client.NewCountriesClient(client.ClientConfig{
		Timeout: clientTimeout,
		Logger:  logger,
	})

	// Service
	countryService := service.NewCountryService(service.ServiceConfig{
		Client: countriesClient,
		Cache:  inMemoryCache,
		Logger: logger,
	})

	// Handler
	countryHandler := handler.NewCountryHandler(countryService, logger)

	// Router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/api/countries/search", countryHandler.Search)

	// HTTP server
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Printf("Server started on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited cleanly")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status": "healthy"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
