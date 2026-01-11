package websocket_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/internal/websocket"
	"connect4-multiplayer/pkg/models"
)

// MockGameServiceIntegration for integration testing
type MockGameServiceIntegration struct {
	mock.Mock
}

func (m *MockGameServiceIntegration) CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	args := m.Called(ctx, player1, player2)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameServiceIntegration) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

// Implement other required methods as no-ops for this test
func (m *MockGameServiceIntegration) GetSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameServiceIntegration) EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error {
	args := m.Called(ctx, gameID, winner, reason)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error) {
	args := m.Called(ctx, gameID)
	return args.String(0), models.PlayerColor(args.String(1)), args.Error(2)
}

func (m *MockGameServiceIntegration) SwitchTurn(ctx context.Context, gameID string) error {
	args := m.Called(ctx, gameID)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error) {
	args := m.Called(ctx, gameID)
	return args.Get(0).(map[string]models.PlayerColor), args.Error(1)
}

func (m *MockGameServiceIntegration) CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error {
	args := m.Called(ctx, gameID, winner)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) GetActiveSessions(ctx context.Context) ([]*models.GameSession, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameServiceIntegration) GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	args := m.Called(ctx, username)
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameServiceIntegration) GetActiveSessionCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGameServiceIntegration) CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error) {
	args := m.Called(ctx, timeout)
	return args.Int(0), args.Error(1)
}

func (m *MockGameServiceIntegration) MarkSessionAbandoned(ctx context.Context, gameID string) error {
	args := m.Called(ctx, gameID)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func (m *MockGameServiceIntegration) StopCleanupWorker() {
	m.Called()
}

func (m *MockGameServiceIntegration) MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) MarkPlayerReconnected(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) GetDisconnectedPlayers(gameID string) map[string]time.Time {
	args := m.Called(gameID)
	return args.Get(0).(map[string]time.Time)
}

func (m *MockGameServiceIntegration) HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error {
	args := m.Called(ctx, gameID, username)
	return args.Error(0)
}

func (m *MockGameServiceIntegration) CacheSession(session *models.GameSession) {
	m.Called(session)
}

func (m *MockGameServiceIntegration) GetCachedSession(gameID string) (*models.GameSession, bool) {
	args := m.Called(gameID)
	return args.Get(0).(*models.GameSession), args.Bool(1)
}

func (m *MockGameServiceIntegration) InvalidateCache(gameID string) {
	m.Called(gameID)
}

func (m *MockGameServiceIntegration) CleanupCache(maxAge time.Duration) int {
	args := m.Called(maxAge)
	return args.Int(0)
}

func (m *MockGameServiceIntegration) GetCacheStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

// Custom room methods
func (m *MockGameServiceIntegration) CreateCustomRoom(ctx context.Context, creator string) (*models.GameSession, string, error) {
	args := m.Called(ctx, creator)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.GameSession), args.String(1), args.Error(2)
}

func (m *MockGameServiceIntegration) JoinCustomRoom(ctx context.Context, roomCode, username string) (*models.GameSession, error) {
	args := m.Called(ctx, roomCode, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameServiceIntegration) GetSessionByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	args := m.Called(ctx, roomCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func TestWebSocketMatchmakingIntegration(t *testing.T) {
	// Create mock game service
	mockGameService := new(MockGameServiceIntegration)

	// Create matchmaking service
	matchmakingConfig := &matchmaking.ServiceConfig{
		MatchTimeout:  1 * time.Second, // Short timeout for test
		MatchInterval: 100 * time.Millisecond,
		Logger:        slog.Default(),
	}
	matchmakingService := matchmaking.NewMatchmakingService(mockGameService, matchmakingConfig)

	// Create WebSocket service with matchmaking
	wsService := websocket.NewService(mockGameService, matchmakingService)

	// Start services
	ctx := context.Background()
	err := wsService.Start(ctx)
	assert.NoError(t, err)

	defer wsService.Stop()

	// Verify services are running
	assert.NotNil(t, wsService.GetHub())
	assert.Equal(t, 0, wsService.GetConnectionCount())

	// Test that matchmaking service is integrated
	// (This is a basic integration test - full WebSocket testing would require actual connections)
	time.Sleep(100 * time.Millisecond) // Allow services to start

	// The integration is successful if no errors occurred during startup
}

func TestWebSocketServiceLifecycle(t *testing.T) {
	// Create mock game service
	mockGameService := new(MockGameServiceIntegration)

	// Create matchmaking service
	matchmakingService := matchmaking.NewMatchmakingService(mockGameService, matchmaking.DefaultServiceConfig())

	// Create WebSocket service
	wsService := websocket.NewService(mockGameService, matchmakingService)

	// Test start
	ctx := context.Background()
	err := wsService.Start(ctx)
	assert.NoError(t, err)

	// Verify service is running
	assert.NotNil(t, wsService.GetHub())
	assert.NotNil(t, wsService.GetWebSocketHandler())

	// Test stop
	err = wsService.Stop()
	assert.NoError(t, err)
}
