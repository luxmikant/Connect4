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

// MockGameService is defined in mocks_test.go

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
	entry1, err := suite.service.JoinQueue(suite.ctx, "player1")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), entry1)

	// Try to join again - should return existing entry (graceful handling)
	entry2, err := suite.service.JoinQueue(suite.ctx, "player1")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), entry2)
	assert.Equal(suite.T(), entry1.Username, entry2.Username)

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
	// Test leaving queue when not in queue - should be gracefully handled
	err := suite.service.LeaveQueue(suite.ctx, "player1")

	// Should not return error (graceful handling)
	assert.NoError(suite.T(), err)
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
