package matchmaking_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"connect4-multiplayer/pkg/models"
)

// MockGameService is a mock implementation of game.GameService
type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	args := m.Called(ctx, player1, player2)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) GetSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error {
	args := m.Called(ctx, gameID, winner, reason)
	return args.Error(0)
}

func (m *MockGameService) GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error) {
	args := m.Called(ctx, gameID)
	return args.String(0), models.PlayerColor(args.String(1)), args.Error(2)
}

func (m *MockGameService) SwitchTurn(ctx context.Context, gameID string) error {
	args := m.Called(ctx, gameID)
	return args.Error(0)
}

func (m *MockGameService) AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(map[string]models.PlayerColor), args.Error(1)
}

func (m *MockGameService) CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error {
	args := m.Called(ctx, gameID, winner)
	return args.Error(0)
}

func (m *MockGameService) GetActiveSessions(ctx context.Context) ([]*models.GameSession, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameService) GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	args := m.Called(ctx, username)
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameService) GetActiveSessionCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGameService) CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error) {
	args := m.Called(ctx, timeout)
	return args.Int(0), args.Error(1)
}

func (m *MockGameService) MarkSessionAbandoned(ctx context.Context, gameID string) error {
	args := m.Called(ctx, gameID)
	return args.Error(0)
}

func (m *MockGameService) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func (m *MockGameService) StopCleanupWorker() {
	m.Called()
}

func (m *MockGameService) MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameService) MarkPlayerReconnected(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameService) GetDisconnectedPlayers(gameID string) map[string]time.Time {
	args := m.Called(gameID)
	return args.Get(0).(map[string]time.Time)
}

func (m *MockGameService) HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameService) CacheSession(session *models.GameSession) {
	m.Called(session)
}

func (m *MockGameService) GetCachedSession(gameID string) (*models.GameSession, bool) {
	args := m.Called(gameID)
	return args.Get(0).(*models.GameSession), args.Bool(1)
}

func (m *MockGameService) InvalidateCache(gameID string) {
	m.Called(gameID)
}

func (m *MockGameService) CleanupCache(maxAge time.Duration) int {
	args := m.Called(maxAge)
	return args.Int(0)
}

func (m *MockGameService) GetCacheStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// Custom room methods
func (m *MockGameService) CreateCustomRoom(ctx context.Context, creator string) (*models.GameSession, string, error) {
	args := m.Called(ctx, creator)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.GameSession), args.String(1), args.Error(2)
}

func (m *MockGameService) JoinCustomRoom(ctx context.Context, roomCode, username string) (*models.GameSession, error) {
	args := m.Called(ctx, roomCode, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) GetSessionByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	args := m.Called(ctx, roomCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameService) RematchCustomRoom(ctx context.Context, gameID, username string) (*models.GameSession, error) {
	args := m.Called(ctx, gameID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}
