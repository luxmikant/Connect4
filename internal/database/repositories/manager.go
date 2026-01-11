package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Manager provides access to all repositories with health checks and connection management
type Manager struct {
	db *gorm.DB

	// Repository instances
	Player      PlayerRepository
	GameSession GameSessionRepository
	PlayerStats PlayerStatsRepository
	Move        MoveRepository
	GameEvent   GameEventRepository
}

// NewManager creates a new repository manager with all repositories
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:          db,
		Player:      NewPlayerRepository(db),
		GameSession: NewGameSessionRepository(db),
		PlayerStats: NewPlayerStatsRepository(db),
		Move:        NewMoveRepository(db),
		GameEvent:   NewGameEventRepository(db),
	}
}

// HealthCheck performs a health check on the database connection
func (m *Manager) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get underlying sql.DB
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Ping the database
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := m.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}

// GetConnectionStats returns database connection statistics
func (m *Manager) GetConnectionStats() (map[string]interface{}, error) {
	sqlDB, err := m.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// Close closes all database connections
func (m *Manager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// BeginTransaction starts a new database transaction
func (m *Manager) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	return m.db.WithContext(ctx).Begin(), nil
}

// WithTransaction executes a function within a database transaction
func (m *Manager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return m.db.WithContext(ctx).Transaction(fn)
}