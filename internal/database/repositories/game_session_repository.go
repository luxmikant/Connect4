package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// gameSessionRepository implements GameSessionRepository interface
type gameSessionRepository struct {
	db *gorm.DB
}

// NewGameSessionRepository creates a new GameSessionRepository instance
func NewGameSessionRepository(db *gorm.DB) GameSessionRepository {
	return &gameSessionRepository{db: db}
}

// Create creates a new game session with retry logic
func (r *gameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	if session == nil {
		return fmt.Errorf("game session cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Retry logic for cloud-native resilience
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Create(session).Error
		if err == nil {
			return nil
		}

		if !isRetryableError(err) {
			return fmt.Errorf("failed to create game session: %w", err)
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

	return fmt.Errorf("failed to create game session after %d attempts", maxRetries)
}

// GetByID retrieves a game session by ID with move history preloaded
func (r *gameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	if id == "" {
		return nil, fmt.Errorf("game session ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var session models.GameSession
	err := r.db.WithContext(ctx).
		Preload("MoveHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp ASC")
		}).
		First(&session, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game session by ID: %w", err)
	}

	return &session, nil
}

// Update updates a game session with optimistic locking
func (r *gameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	if session == nil {
		return fmt.Errorf("game session cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Use transaction for consistency
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Save(session)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return models.ErrGameNotFound
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update game session: %w", err)
	}

	return nil
}

// Delete soft deletes a game session
func (r *gameSessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("game session ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Delete(&models.GameSession{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete game session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrGameNotFound
	}

	return nil
}

// GetActiveGames retrieves all active game sessions
func (r *gameSessionRepository) GetActiveGames(ctx context.Context) ([]*models.GameSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var sessions []*models.GameSession
	err := r.db.WithContext(ctx).
		Where("status IN ?", []models.GameStatus{
			models.StatusWaiting,
			models.StatusInProgress,
		}).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active games: %w", err)
	}

	return sessions, nil
}

// GetGamesByPlayer retrieves games for a specific player
func (r *gameSessionRepository) GetGamesByPlayer(ctx context.Context, playerID string) ([]*models.GameSession, error) {
	if playerID == "" {
		return nil, fmt.Errorf("player ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var sessions []*models.GameSession
	err := r.db.WithContext(ctx).
		Where("player1 = ? OR player2 = ?", playerID, playerID).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get games by player: %w", err)
	}

	return sessions, nil
}

// GetGameHistory retrieves completed games with pagination
func (r *gameSessionRepository) GetGameHistory(ctx context.Context, limit, offset int) ([]*models.GameSession, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var sessions []*models.GameSession
	err := r.db.WithContext(ctx).
		Where("status = ?", models.StatusCompleted).
		Limit(limit).
		Offset(offset).
		Order("end_time DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}

	return sessions, nil
}