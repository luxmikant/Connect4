package main

import (
	"flag"
	"log"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
)

func main() {
	var direction = flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, _, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	migrator := database.NewMigrator(db)
	
	switch *direction {
	case "up":
		log.Println("Running migrations...")
		if err := migrator.Up(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
	case "down":
		log.Println("Rolling back migrations...")
		if err := migrator.Down(); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		log.Println("Rollback completed successfully")
	default:
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", *direction)
	}
}