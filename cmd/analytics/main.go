package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create analytics service
	analyticsService, err := analytics.NewService(cfg.Kafka, db)
	if err != nil {
		log.Fatalf("Failed to create analytics service: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start analytics service
	go func() {
		log.Println("Starting analytics service...")
		if err := analyticsService.Start(ctx); err != nil {
			log.Fatalf("Analytics service failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down analytics service...")

	// Cancel context to stop service
	cancel()

	log.Println("Analytics service exited")
}