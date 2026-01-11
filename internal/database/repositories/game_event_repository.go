package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// gameEventRepository implements GameEventRepository interface
type gameEventRepository struct {
	db *gorm.DB
}

// NewGameEventRepository creates a new GameEventRepository instance
func NewGameEventRepository(db *gorm.DB) GameEventRepository {
	return &gameEventRepository{db: db}
}

// Create creates a new game event with retry logic
func (r *gameEventRepository) Create(ctx context.Context, event *models.GameEvent) error {
	if event == nil {
		return fmt.Errorf("game event cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Retry logic for cloud-native resilience
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Create(event).Error
		if err == nil {
			return nil
		}

		if !isRetryableError(err) {
			return fmt.Errorf("failed to create game event: %w", err)
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

	return fmt.Errorf("failed to create game event after %d attempts", maxRetries)
}

// GetByID retrieves a game event by ID
func (r *gameEventRepository) GetByID(ctx context.Context, id string) (*models.GameEvent, error) {
	if id == "" {
		return nil, fmt.Errorf("game event ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var event models.GameEvent
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("game event not found")
		}
		return nil, fmt.Errorf("failed to get game event by ID: %w", err)
	}

	return &event, nil
}

// GetByGameID retrieves all events for a specific game
func (r *gameEventRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.GameEvent, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var events []*models.GameEvent
	err := r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Order("timestamp ASC").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get events by game ID: %w", err)
	}

	return events, nil
}

// GetByEventType retrieves events by type with pagination
func (r *gameEventRepository) GetByEventType(ctx context.Context, eventType models.EventType, limit, offset int) ([]*models.GameEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var events []*models.GameEvent
	err := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}

	return events, nil
}

// GetEventsByTimeRange retrieves events within a time range with pagination
func (r *gameEventRepository) GetEventsByTimeRange(ctx context.Context, start, end string, limit, offset int) ([]*models.GameEvent, error) {
	if start == "" || end == "" {
		return nil, fmt.Errorf("start and end time cannot be empty")
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var events []*models.GameEvent
	err := r.db.WithContext(ctx).
		Where("timestamp BETWEEN ? AND ?", start, end).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get events by time range: %w", err)
	}

	return events, nil
}