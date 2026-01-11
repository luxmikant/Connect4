package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// playerStatsRepository implements PlayerStatsRepository interface
type playerStatsRepository struct {
	db *gorm.DB
}

// NewPlayerStatsRepository creates a new PlayerStatsRepository instance
func NewPlayerStatsRepository(db *gorm.DB) PlayerStatsRepository {
	return &playerStatsRepository{db: db}
}

// Create creates new player statistics with retry logic
func (r *playerStatsRepository) Create(ctx context.Context, stats *models.PlayerStats) error {
	if stats == nil {
		return fmt.Errorf("player stats cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Retry logic for cloud-native resilience
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Create(stats).Error
		if err == nil {
			return nil
		}

		if !isRetryableError(err) {
			return fmt.Errorf("failed to create player stats: %w", err)
		}

		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffDuration):
				continue
			}
		}
	}

	return fmt.Errorf("failed to create player stats after %d attempts", maxRetries)
}

// GetByID retrieves player statistics by ID
func (r *playerStatsRepository) GetByID(ctx context.Context, id string) (*models.PlayerStats, error) {
	if id == "" {
		return nil, fmt.Errorf("player stats ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var stats models.PlayerStats
	err := r.db.WithContext(ctx).First(&stats, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player stats by ID: %w", err)
	}

	return &stats, nil
}

// GetByUsername retrieves player statistics by username
func (r *playerStatsRepository) GetByUsername(ctx context.Context, username string) (*models.PlayerStats, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var stats models.PlayerStats
	err := r.db.WithContext(ctx).First(&stats, "username = ?", username).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player stats by username: %w", err)
	}

	return &stats, nil
}

// Update updates player statistics with optimistic locking
func (r *playerStatsRepository) Update(ctx context.Context, stats *models.PlayerStats) error {
	if stats == nil {
		return fmt.Errorf("player stats cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Save(stats)
	if result.Error != nil {
		return fmt.Errorf("failed to update player stats: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPlayerNotFound
	}

	return nil
}

// Delete soft deletes player statistics
func (r *playerStatsRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("player stats ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Delete(&models.PlayerStats{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete player stats: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPlayerNotFound
	}

	return nil
}

// GetLeaderboard retrieves top players sorted by wins
func (r *playerStatsRepository) GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error) {
	if limit <= 0 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var stats []*models.PlayerStats
	err := r.db.WithContext(ctx).
		Where("games_played > 0").
		Order("games_won DESC, win_rate DESC, games_played DESC").
		Limit(limit).
		Find(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return stats, nil
}

// UpdateGameStats updates player statistics after a game with atomic operations
func (r *playerStatsRepository) UpdateGameStats(ctx context.Context, username string, won bool, gameDuration int) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use transaction for atomic updates
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get or create player stats
		var stats models.PlayerStats
		err := tx.Where("username = ?", username).First(&stats).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new stats record
				stats = models.PlayerStats{
					Username:    username,
					GamesPlayed: 0,
					GamesWon:    0,
					WinRate:     0.0,
					AvgGameTime: 0,
					LastPlayed:  time.Now(),
				}
			} else {
				return fmt.Errorf("failed to get player stats: %w", err)
			}
		}

		// Update statistics
		stats.UpdateGameStats(won, gameDuration)

		// Save updated stats
		if err := tx.Save(&stats).Error; err != nil {
			return fmt.Errorf("failed to save updated stats: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update game stats: %w", err)
	}

	return nil
}