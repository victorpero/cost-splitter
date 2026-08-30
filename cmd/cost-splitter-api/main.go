package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/victorpero/cost-splitter/internal/api"
	"github.com/victorpero/cost-splitter/internal/application"
)

func main() {
	addr := envOrDefault("API_ADDR", "127.0.0.1:8080")
	currency := envOrDefault("API_CURRENCY", "SEK")
	allowedOrigins := splitList(envOrDefault("API_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"))

	handler := api.NewServer(api.Config{
		Currency:       currency,
		AllowedOrigins: allowedOrigins,
	}, application.NewService())
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       35 * time.Second,
		WriteTimeout:      35 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownSignals, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownSignals.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("API shutdown: %v", err)
		}
	}()

	log.Printf("cost-splitter API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
