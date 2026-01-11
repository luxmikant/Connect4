package repositories

import (
	"context"

	"connect4-multiplayer/pkg/models"
)

// PlayerRepository defines the interface for player data operations
type PlayerRepository interface {
	Create(ctx context.Context, player *models.Player) error
	GetByID(ctx context.Context, id string) (*models.Player, error)
	GetByUsername(ctx context.Context, username string) (*models.Player, error)
	Update(ctx context.Context, player *models.Player) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Player, error)
}

// GameSessionRepository defines the interface for game session data operations
type GameSessionRepository interface {
	Create(ctx context.Context, session *models.GameSession) error
	GetByID(ctx context.Context, id string) (*models.GameSession, error)
	Update(ctx context.Context, session *models.GameSession) error
	Delete(ctx context.Context, id string) error
	GetActiveGames(ctx context.Context) ([]*models.GameSession, error)
	GetGamesByPlayer(ctx context.Context, playerID string) ([]*models.GameSession, error)
	GetGameHistory(ctx context.Context, limit, offset int) ([]*models.GameSession, error)
}

// PlayerStatsRepository defines the interface for player statistics operations
type PlayerStatsRepository interface {
	Create(ctx context.Context, stats *models.PlayerStats) error
	GetByID(ctx context.Context, id string) (*models.PlayerStats, error)
	GetByUsername(ctx context.Context, username string) (*models.PlayerStats, error)
	Update(ctx context.Context, stats *models.PlayerStats) error
	Delete(ctx context.Context, id string) error
	GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error)
	UpdateGameStats(ctx context.Context, username string, won bool, gameDuration int) error
}

// MoveRepository defines the interface for move data operations
type MoveRepository interface {
	Create(ctx context.Context, move *models.Move) error
	GetByID(ctx context.Context, id string) (*models.Move, error)
	GetByGameID(ctx context.Context, gameID string) ([]*models.Move, error)
	Delete(ctx context.Context, id string) error
	GetMoveHistory(ctx context.Context, gameID string, limit int) ([]*models.Move, error)
}

// GameEventRepository defines the interface for analytics event operations
type GameEventRepository interface {
	Create(ctx context.Context, event *models.GameEvent) error
	GetByID(ctx context.Context, id string) (*models.GameEvent, error)
	GetByGameID(ctx context.Context, gameID string) ([]*models.GameEvent, error)
	GetByEventType(ctx context.Context, eventType models.EventType, limit, offset int) ([]*models.GameEvent, error)
	GetEventsByTimeRange(ctx context.Context, start, end string, limit, offset int) ([]*models.GameEvent, error)
}