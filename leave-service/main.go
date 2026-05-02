package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asdlc-repos/mekanamwedakarannaone2/leave-service/internal/db"
	"github.com/asdlc-repos/mekanamwedakarannaone2/leave-service/internal/handlers"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/leavedb?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	database, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.MigrateSchema(database); err != nil {
		log.Fatalf("failed to migrate schema: %v", err)
	}

	if err := db.SeedData(database); err != nil {
		log.Printf("warning: failed to seed data: %v", err)
	}

	h := handlers.New(database)
	mux := http.NewServeMux()
	handler := h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("leave-service listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited gracefully")
}
