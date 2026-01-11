package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

// Initialize initializes the database connection and returns repository manager
func Initialize(cfg config.DatabaseConfig) (*gorm.DB, *repositories.Manager, error) {
	// Configure GORM logger
	var logLevel logger.LogLevel
	switch cfg.SSLMode {
	case "production":
		logLevel = logger.Error
	default:
		logLevel = logger.Info
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for cloud-native deployment
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create repository manager
	repoManager := repositories.NewManager(db)

	return db, repoManager, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Player{},
		&models.GameSession{},
		&models.Move{},
		&models.PlayerStats{},
		&models.GameEvent{},
	)
}