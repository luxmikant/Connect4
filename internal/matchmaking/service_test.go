package matchmaking_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"connect4-multiplayer/internal/matchmaking"
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

// Implement other required methods as no-ops for this test
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

// MatchmakingServiceTestSuite contains tests for the matchmaking service
type MatchmakingServiceTestSuite struct {
	suite.Suite
	mockGameService *MockGameService
	service         matchmaking.MatchmakingService
	ctx             context.Context
}

func (suite *MatchmakingServiceTestSuite) SetupTest() {
	suite.mockGameService = new(MockGameService)
	suite.ctx = context.Background()
	
	config := &matchmaking.ServiceConfig{
		MatchTimeout:  2 * time.Second, // Shorter timeout for tests
		MatchInterval: 100 * time.Millisecond,
		Logger:        slog.Default(), // Use default logger instead of nil
	}
	
	suite.service = matchmaking.NewMatchmakingService(suite.mockGameService, config)
}

func (suite *MatchmakingServiceTestSuite) TearDown() {
	suite.service.StopMatchmaking()
}

func (suite *MatchmakingServiceTestSuite) TestJoinQueue_Success() {
	// Mock: player not in active game
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	
	// Test joining queue
	entry, err := suite.service.JoinQueue(suite.ctx, "player1")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), entry)
	assert.Equal(suite.T(), "player1", entry.Username)
	assert.False(suite.T(), entry.JoinedAt.IsZero())
	assert.False(suite.T(), entry.Timeout.IsZero())
	
	// Verify queue length
	length := suite.service.GetQueueLength(suite.ctx)
	assert.Equal(suite.T(), 1, length)
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestJoinQueue_EmptyUsername() {
	// Test joining queue with empty username
	entry, err := suite.service.JoinQueue(suite.ctx, "")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), entry)
	assert.Contains(suite.T(), err.Error(), "username cannot be empty")
}

func (suite *MatchmakingServiceTestSuite) TestJoinQueue_PlayerAlreadyInActiveGame() {
	// Mock: player already in active game
	activeSession := &models.GameSession{
		ID:     "game-123",
		Status: models.StatusInProgress,
	}
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(activeSession, nil)
	
	// Test joining queue
	entry, err := suite.service.JoinQueue(suite.ctx, "player1")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), entry)
	assert.Contains(suite.T(), err.Error(), "already in an active game")
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestJoinQueue_PlayerAlreadyInQueue() {
	// Mock: player not in active game
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	
	// Join queue first time
	_, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	// Try to join again
	entry, err := suite.service.JoinQueue(suite.ctx, "player1")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), entry)
	assert.Contains(suite.T(), err.Error(), "already in queue")
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestLeaveQueue_Success() {
	// Mock: player not in active game
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	
	// Join queue first
	_, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	// Leave queue
	err = suite.service.LeaveQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	// Verify queue is empty
	length := suite.service.GetQueueLength(suite.ctx)
	assert.Equal(suite.T(), 0, length)
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestLeaveQueue_PlayerNotInQueue() {
	// Test leaving queue when not in queue
	err := suite.service.LeaveQueue(suite.ctx, "player1")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not in queue")
}

func (suite *MatchmakingServiceTestSuite) TestGetQueueStatus_InQueue() {
	// Mock: player not in active game
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	
	// Join queue
	_, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	// Get status
	status, err := suite.service.GetQueueStatus(suite.ctx, "player1")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.True(suite.T(), status.InQueue)
	assert.Equal(suite.T(), 1, status.Position)
	assert.True(suite.T(), status.WaitTime >= 0)
	assert.True(suite.T(), status.TimeRemaining > 0)
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestGetQueueStatus_NotInQueue() {
	// Get status for player not in queue
	status, err := suite.service.GetQueueStatus(suite.ctx, "player1")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.False(suite.T(), status.InQueue)
}

func (suite *MatchmakingServiceTestSuite) TestPlayerMatchmaking_TwoPlayers() {
	// Mock: players not in active games
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player2").Return(nil, assert.AnError)
	
	// Mock: game creation
	gameSession := &models.GameSession{
		ID:      "game-123",
		Player1: "player1",
		Player2: "player2",
		Status:  models.StatusInProgress,
	}
	suite.mockGameService.On("CreateSession", mock.AnythingOfType("*context.cancelCtx"), "player1", "player2").Return(gameSession, nil)
	
	// Set up callback to track game creation
	var createdGame *models.GameSession
	suite.service.SetGameCreatedCallback(func(ctx context.Context, player1, player2 string, gameSession *models.GameSession) error {
		createdGame = gameSession
		return nil
	})
	
	// Join both players to queue
	_, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	_, err = suite.service.JoinQueue(suite.ctx, "player2")
	assert.NoError(suite.T(), err)
	
	// Start matchmaking
	err = suite.service.StartMatchmaking(suite.ctx)
	assert.NoError(suite.T(), err)
	
	// Wait for matchmaking to process
	time.Sleep(200 * time.Millisecond)
	
	// Stop matchmaking
	suite.service.StopMatchmaking()
	
	// Verify game was created
	assert.NotNil(suite.T(), createdGame)
	assert.Equal(suite.T(), "game-123", createdGame.ID)
	
	// Verify queue is empty
	length := suite.service.GetQueueLength(suite.ctx)
	assert.Equal(suite.T(), 0, length)
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func (suite *MatchmakingServiceTestSuite) TestBotMatchmaking_Timeout() {
	// Mock: player not in active game
	suite.mockGameService.On("GetActiveSessionByPlayer", suite.ctx, "player1").Return(nil, assert.AnError)
	
	// Mock: bot game creation
	gameSession := &models.GameSession{
		ID:      "game-456",
		Player1: "player1",
		Player2: "bot_123456", // Mock bot name
		Status:  models.StatusInProgress,
	}
	suite.mockGameService.On("CreateSession", mock.AnythingOfType("*context.cancelCtx"), "player1", mock.AnythingOfType("string")).Return(gameSession, nil)
	
	// Set up callback to track bot game creation
	var createdBotGame *models.GameSession
	suite.service.SetBotGameCallback(func(ctx context.Context, player string, gameSession *models.GameSession) error {
		createdBotGame = gameSession
		return nil
	})
	
	// Join player to queue
	_, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	
	// Start matchmaking
	err = suite.service.StartMatchmaking(suite.ctx)
	assert.NoError(suite.T(), err)
	
	// Wait for timeout and processing
	time.Sleep(2500 * time.Millisecond) // Wait longer than 2-second timeout
	
	// Stop matchmaking
	suite.service.StopMatchmaking()
	
	// Verify bot game was created
	assert.NotNil(suite.T(), createdBotGame)
	assert.Equal(suite.T(), "game-456", createdBotGame.ID)
	
	// Verify queue is empty
	length := suite.service.GetQueueLength(suite.ctx)
	assert.Equal(suite.T(), 0, length)
	
	suite.mockGameService.AssertExpectations(suite.T())
}

func TestMatchmakingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MatchmakingServiceTestSuite))
}