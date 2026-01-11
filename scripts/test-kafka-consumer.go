package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
)

func main() {
	fmt.Println("🔍 Testing Kafka Consumer (Analytics Service)")
	fmt.Println(strings.Repeat("=", 60))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize database
	fmt.Println("🗄️  Connecting to database...")
	db, _, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	fmt.Println("✅ Database connected")

	// Create analytics service
	fmt.Println("🔧 Creating analytics service...")
	analyticsService, err := analytics.NewService(cfg.Kafka, db)
	if err != nil {
		log.Fatalf("❌ Failed to create analytics service: %v", err)
	}
	fmt.Println("✅ Analytics service created")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start analytics service in background
	fmt.Println("🚀 Starting analytics service...")
	fmt.Println("📋 Configuration:")
	fmt.Printf("   Topic: %s\n", cfg.Kafka.Topic)
	fmt.Printf("   Consumer Group: %s\n", cfg.Kafka.ConsumerGroup)
	fmt.Printf("   Bootstrap: %s\n", cfg.Kafka.BootstrapServers)
	fmt.Println()
	fmt.Println("⏳ Waiting for messages... (Press Ctrl+C to stop)")
	fmt.Println("💡 Run the producer test in another terminal: go run scripts/test-kafka-cloud.go")
	fmt.Println()

	// Start service
	go func() {
		if err := analyticsService.Start(ctx); err != nil {
			log.Printf("❌ Analytics service error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Show status every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	
	for {
		select {
		case <-quit:
			fmt.Println("\n🛑 Shutting down analytics service...")
			cancel()
			time.Sleep(2 * time.Second) // Give time for graceful shutdown
			fmt.Println("✅ Analytics service stopped")
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			fmt.Printf("⏰ Running for %v - Still listening for messages...\n", elapsed.Round(time.Second))
		}
	}
}