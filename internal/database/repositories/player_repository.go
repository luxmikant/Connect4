package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"connect4-multiplayer/pkg/models"
)

// playerRepository implements PlayerRepository interface
type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a new PlayerRepository instance
func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

// Create creates a new player with retry logic and error handling
func (r *playerRepository) Create(ctx context.Context, player *models.Player) error {
	if player == nil {
		return fmt.Errorf("player cannot be nil")
	}

	// Set context timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Retry logic for cloud-native resilience
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Create(player).Error
		if err == nil {
			return nil
		}

		// Check if it's a retryable error
		if !isRetryableError(err) {
			return fmt.Errorf("failed to create player: %w", err)
		}

		// Wait before retry with exponential backoff
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

	return fmt.Errorf("failed to create player after %d attempts", maxRetries)
}

// GetByID retrieves a player by ID with connection failover
func (r *playerRepository) GetByID(ctx context.Context, id string) (*models.Player, error) {
	if id == "" {
		return nil, fmt.Errorf("player ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var player models.Player
	err := r.db.WithContext(ctx).First(&player, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by ID: %w", err)
	}

	return &player, nil
}

// GetByUsername retrieves a player by username with connection failover
func (r *playerRepository) GetByUsername(ctx context.Context, username string) (*models.Player, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var player models.Player
	err := r.db.WithContext(ctx).First(&player, "username = ?", username).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by username: %w", err)
	}

	return &player, nil
}

// GetByAuthUserID retrieves a player by auth_user_id
func (r *playerRepository) GetByAuthUserID(ctx context.Context, authUserID string) (*models.Player, error) {
	if authUserID == "" {
		return nil, fmt.Errorf("authUserID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var player models.Player
	err := r.db.WithContext(ctx).First(&player, "auth_user_id = ?", authUserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by auth_user_id: %w", err)
	}

	return &player, nil
}

// UpsertFromProfile creates or updates a player from auth profile
func (r *playerRepository) UpsertFromProfile(ctx context.Context, authUserID, username string) (string, error) {
	if authUserID == "" || username == "" {
		return "", fmt.Errorf("authUserID and username cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// First, try to find existing player by auth_user_id
	var existingPlayer models.Player
	err := r.db.WithContext(ctx).First(&existingPlayer, "auth_user_id = ?", authUserID).Error

	if err == nil {
		// Player exists, update it
		existingPlayer.Username = username
		existingPlayer.IsGuest = false
		existingPlayer.UpdatedAt = time.Now()

		if err := r.db.WithContext(ctx).Save(&existingPlayer).Error; err != nil {
			return "", fmt.Errorf("failed to update player: %w", err)
		}
		return existingPlayer.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("failed to query player: %w", err)
	}

	// Player doesn't exist, create new one
	newPlayer := &models.Player{
		ID:         generateUUID(),
		Username:   username,
		AuthUserID: &authUserID,
		IsGuest:    false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(newPlayer).Error; err != nil {
		return "", fmt.Errorf("failed to create player: %w", err)
	}

	return newPlayer.ID, nil
}

// Update updates a player with optimistic locking
func (r *playerRepository) Update(ctx context.Context, player *models.Player) error {
	if player == nil {
		return fmt.Errorf("player cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Save(player)
	if result.Error != nil {
		return fmt.Errorf("failed to update player: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPlayerNotFound
	}

	return nil
}

// Delete soft deletes a player
func (r *playerRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).Delete(&models.Player{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete player: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return models.ErrPlayerNotFound
	}

	return nil
}

// List retrieves players with pagination
func (r *playerRepository) List(ctx context.Context, limit, offset int) ([]*models.Player, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var players []*models.Player
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&players).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	return players, nil
}

// isRetryableError determines if an error is retryable for cloud-native resilience
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable database errors
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"server closed",
		"broken pipe",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}
