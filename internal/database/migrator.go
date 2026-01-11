package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// Migrator handles database migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Up runs all migrations
func (m *Migrator) Up() error {
	// Use GORM AutoMigrate to ensure schema is up to date
	// This handles all column additions, type changes, and index creation automatically
	if err := m.db.AutoMigrate(
		&models.Player{},
		&models.GameSession{},
		&models.Move{},
		&models.PlayerStats{},
		&models.GameEvent{},
		&models.AnalyticsSnapshot{},
	); err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	// Try to run SQL migrations if directory exists, but don't fail if it doesn't
	_ = m.runSQLMigrations()

	return nil
}

// runSQLMigrations runs SQL migration files from the migrations directory
func (m *Migrator) runSQLMigrations() error {
	migrationsDir := "migrations"

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// No migrations directory, skip SQL migrations
		return nil
	}

	// Read all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort SQL files
	var sqlFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	// Sort files by name to ensure proper order
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	// Execute each migration file
	for _, file := range sqlFiles {
		if err := m.executeSQLFile(filepath.Join(migrationsDir, file.Name())); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// executeSQLFile executes a SQL migration file
func (m *Migrator) executeSQLFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	// Execute the SQL content
	if err := m.db.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("failed to execute SQL from %s: %w", filePath, err)
	}

	return nil
}

// Down rolls back migrations (drops all tables)
func (m *Migrator) Down() error {
	// Drop tables in reverse order to handle foreign key constraints
	tables := []interface{}{
		&models.GameEvent{},
		&models.PlayerStats{},
		&models.Move{},
		&models.GameSession{},
		&models.Player{},
	}

	for _, table := range tables {
		if err := m.db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	return nil
}

// createIndexes creates additional database indexes for better performance
// Note: Most indexes are now created via SQL migration files
func (m *Migrator) createIndexes() error {
	// This method is kept for any additional indexes that might be needed
	// beyond what's defined in the SQL migration files

	// Example: Create any additional composite indexes if needed
	// Most indexes are now handled in the SQL migration files

	return nil
}
