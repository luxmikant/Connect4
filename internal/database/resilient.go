package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database/repositories"
)

// ResilientDB wraps a database connection with retry logic and health status.
// It allows the server to start even when the database is temporarily unavailable,
// reconnecting in the background with exponential backoff.
type ResilientDB struct {
	mu          sync.RWMutex
	db          *gorm.DB
	repoManager *repositories.Manager
	cfg         config.DatabaseConfig
	connected   bool
	lastError   error
	stopRetry   chan struct{}
}

// NewResilientDB creates a ResilientDB that attempts an initial connection,
// and if it fails, retries in the background instead of crashing.
func NewResilientDB(cfg config.DatabaseConfig) *ResilientDB {
	r := &ResilientDB{
		cfg:       cfg,
		stopRetry: make(chan struct{}),
	}

	// Attempt initial connection
	if err := r.connect(); err != nil {
		log.Printf("⚠️  Initial database connection failed: %v", err)
		log.Println("⚠️  Server will start without database — retrying in background...")
		go r.retryLoop()
	} else {
		log.Println("✅ Database connected successfully")
	}

	return r
}

// connect attempts a single database connection.
func (r *ResilientDB) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cfg.URL == "" {
		r.lastError = fmt.Errorf("DATABASE_URL is empty — set it to your Neon/Supabase/external Postgres URL")
		return r.lastError
	}

	var logLevel logger.LogLevel
	switch r.cfg.SSLMode {
	case "production", "require":
		logLevel = logger.Error
	default:
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(r.cfg.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		r.lastError = fmt.Errorf("failed to open database: %w", err)
		return r.lastError
	}

	sqlDB, err := db.DB()
	if err != nil {
		r.lastError = fmt.Errorf("failed to get underlying sql.DB: %w", err)
		return r.lastError
	}

	// Connection pool settings (tuned for serverless DBs like Neon)
	sqlDB.SetMaxOpenConns(r.cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(r.cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(r.cfg.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		r.lastError = fmt.Errorf("failed to ping database: %w", err)
		return r.lastError
	}

	r.db = db
	r.repoManager = repositories.NewManager(db)
	r.connected = true
	r.lastError = nil
	return nil
}

// retryLoop retries database connection with exponential backoff.
func (r *ResilientDB) retryLoop() {
	backoff := 5 * time.Second
	maxBackoff := 2 * time.Minute

	for {
		select {
		case <-r.stopRetry:
			return
		case <-time.After(backoff):
			log.Printf("🔄 Retrying database connection (backoff: %v)...", backoff)
			if err := r.connect(); err != nil {
				log.Printf("⚠️  Retry failed: %v", err)
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else {
				log.Println("✅ Database reconnected successfully!")
				// Run migrations after reconnection
				r.runMigrations()
				return
			}
		}
	}
}

// runMigrations runs database migrations after a successful connection.
func (r *ResilientDB) runMigrations() {
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()

	if db == nil {
		return
	}

	log.Println("Running database migrations after reconnection...")
	migrator := NewMigrator(db)
	if err := migrator.Up(); err != nil {
		log.Printf("⚠️  Post-reconnection migration failed: %v", err)
	} else {
		log.Println("✅ Migrations completed successfully")
	}
}

// IsConnected returns whether the database is currently reachable.
func (r *ResilientDB) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connected
}

// LastError returns the last connection error, if any.
func (r *ResilientDB) LastError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastError
}

// DB returns the underlying *gorm.DB, or nil if not connected.
func (r *ResilientDB) DB() *gorm.DB {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.db
}

// RepoManager returns the repository manager, or nil if not connected.
func (r *ResilientDB) RepoManager() *repositories.Manager {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.repoManager
}

// HealthCheck performs an active database health check.
func (r *ResilientDB) HealthCheck() error {
	r.mu.RLock()
	db := r.db
	connected := r.connected
	r.mu.RUnlock()

	if !connected || db == nil {
		return fmt.Errorf("database not connected")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		// Mark as disconnected and start retry loop
		r.mu.Lock()
		r.connected = false
		r.lastError = err
		r.mu.Unlock()
		go r.retryLoop()
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close cleanly shuts down the database connection.
func (r *ResilientDB) Close() error {
	close(r.stopRetry)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.db != nil {
		sqlDB, err := r.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
