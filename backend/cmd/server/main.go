package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/database"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/httpapi"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")
	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.ApplyMigrations(ctx, pool, os.Getenv("MIGRATIONS_DIR")); err != nil {
		log.Fatal(err)
	}

	handler := httpapi.NewServer(pool).Handler()

	server := &http.Server{Addr: ":" + envOrDefault("APP_PORT", "8080"), Handler: handler}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
