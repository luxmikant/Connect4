//go:build integration
// +build integration

package handlers_test

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

// initializeTestDatabase initializes an in-memory SQLite database for testing
func initializeTestDatabase() (*gorm.DB, *repositories.Manager, error) {
	// Open SQLite in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce noise in tests
	})
	if err != nil {
		return nil, nil, err
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.Player{},
		&models.GameSession{},
		&models.Move{},
		&models.PlayerStats{},
		&models.GameEvent{},
	)
	if err != nil {
		return nil, nil, err
	}

	// Create repository manager
	repoManager := repositories.NewManager(db)

	return db, repoManager, nil
}