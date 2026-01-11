package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// moveRepository implements MoveRepository interface
type moveRepository struct {
	db *gorm.DB
}

// NewMoveRepository creates a new MoveRepository instance
func NewMoveRepository(db *gorm.DB) MoveRepository {
	return &moveRepository{db: db}
}

// Create creates a new move with retry logic
func (r *moveRepository) Create(ctx context.Context, move *models.Move) error {
	if move == nil {
		return fmt.Errorf("move cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Validate move before creating
	if !move.IsValid() {
		return models.ErrInvalidMove
	}

	// Retry logic for cloud-native resilience
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Create(move).Error
		if err == nil {
			return nil
		}

		if !isRetryableError(err) {
			return fmt.Errorf("failed to create move: %w", err)
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

	return fmt.Errorf("failed to create move after %d attempts", maxRetries)
}

// GetByID retrieves a move by ID
func (r *moveRepository) GetByID(ctx context.Context, id string) (*models.Move, error) {
	if id == "" {
		return nil, fmt.Errorf("move ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var move models.Move
	err := r.db.WithContext(ctx).First(&move, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("move not found")
		}
		return nil, fmt.Errorf("failed to get move by ID: %w", err)
	}

	return &move, nil
}

// GetByGameID retrieves all moves for a specific game
func (r *moveRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.Move, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var moves []*models.Move
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Order("timestamp ASC").
		Find(&moves).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get moves by game ID: %w", err)
	}

	return moves, nil
}

// Delete soft deletes a move
func (r *moveRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("move ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Delete(&models.Move{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete move: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("move not found")
	}

	return nil
}

// GetMoveHistory retrieves move history for a game with pagination
func (r *moveRepository) GetMoveHistory(ctx context.Context, gameID string, limit int) ([]*models.Move, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	if limit <= 0 {
		limit = 50 // Default limit for move history
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var moves []*models.Move
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Order("timestamp ASC").
		Limit(limit).
		Find(&moves).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get move history: %w", err)
	}

	return moves, nil
}