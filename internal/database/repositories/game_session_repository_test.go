package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

type GameSessionRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repositories.GameSessionRepository
}

func (suite *GameSessionRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.GameSession{}, &models.Move{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repositories.NewGameSessionRepository(db)
}

func (suite *GameSessionRepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *GameSessionRepositoryTestSuite) TestCreate_Success() {
	ctx := context.Background()
	gameSession := &models.GameSession{
		ID:          "test-game-1",
		Player1:     "player1",
		Player2:     "player2",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusWaiting,
		Board:       models.NewBoard(),
	}

	err := suite.repo.Create(ctx, gameSession)
	assert.NoError(suite.T(), err)

	// Verify game session was created
	var retrieved models.GameSession
	err = suite.db.First(&retrieved, "id = ?", gameSession.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), gameSession.Player1, retrieved.Player1)
	assert.Equal(suite.T(), gameSession.Player2, retrieved.Player2)
}

func (suite *GameSessionRepositoryTestSuite) TestCreate_NilGameSession() {
	ctx := context.Background()
	err := suite.repo.Create(ctx, nil)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "game session cannot be nil")
}

func (suite *GameSessionRepositoryTestSuite) TestGetByID_Success() {
	ctx := context.Background()
	
	// Create test game session
	gameSession := &models.GameSession{
		ID:          "test-game-2",
		Player1:     "player1",
		Player2:     "player2",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		Board:       models.NewBoard(),
	}
	err := suite.db.Create(gameSession).Error
	suite.Require().NoError(err)

	// Retrieve game session
	retrieved, err := suite.repo.GetByID(ctx, gameSession.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), gameSession.ID, retrieved.ID)
	assert.Equal(suite.T(), gameSession.Player1, retrieved.Player1)
	assert.Equal(suite.T(), gameSession.Status, retrieved.Status)
}

func (suite *GameSessionRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	
	retrieved, err := suite.repo.GetByID(ctx, "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
	assert.Equal(suite.T(), models.ErrGameNotFound, err)
}

func (suite *GameSessionRepositoryTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	
	// Create test game session
	gameSession := &models.GameSession{
		ID:          "test-game-3",
		Player1:     "player1",
		Player2:     "player2",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusWaiting,
		Board:       models.NewBoard(),
	}
	err := suite.db.Create(gameSession).Error
	suite.Require().NoError(err)

	// Update game session
	gameSession.Status = models.StatusInProgress
	gameSession.CurrentTurn = models.PlayerColorYellow
	err = suite.repo.Update(ctx, gameSession)
	assert.NoError(suite.T(), err)

	// Verify update
	var retrieved models.GameSession
	err = suite.db.First(&retrieved, "id = ?", gameSession.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.StatusInProgress, retrieved.Status)
	assert.Equal(suite.T(), models.PlayerColorYellow, retrieved.CurrentTurn)
}

func (suite *GameSessionRepositoryTestSuite) TestGetActiveGames_Success() {
	ctx := context.Background()
	
	// Create test game sessions with different statuses
	gameSessions := []*models.GameSession{
		{
			ID:          "test-game-4",
			Player1:     "player1",
			Player2:     "player2",
			CurrentTurn: models.PlayerColorRed,
			Status:      models.StatusWaiting,
			Board:       models.NewBoard(),
		},
		{
			ID:          "test-game-5",
			Player1:     "player3",
			Player2:     "player4",
			CurrentTurn: models.PlayerColorRed,
			Status:      models.StatusInProgress,
			Board:       models.NewBoard(),
		},
		{
			ID:          "test-game-6",
			Player1:     "player5",
			Player2:     "player6",
			CurrentTurn: models.PlayerColorRed,
			Status:      models.StatusCompleted,
			Board:       models.NewBoard(),
		},
	}

	for _, gameSession := range gameSessions {
		err := suite.db.Create(gameSession).Error
		suite.Require().NoError(err)
	}

	// Get active games
	activeGames, err := suite.repo.GetActiveGames(ctx)
	assert.NoError(suite.T(), err)
	
	// Should return only waiting and in-progress games
	activeCount := 0
	for _, game := range activeGames {
		if game.Status == models.StatusWaiting || game.Status == models.StatusInProgress {
			activeCount++
		}
	}
	assert.GreaterOrEqual(suite.T(), activeCount, 2)
}

func (suite *GameSessionRepositoryTestSuite) TestGetGamesByPlayer_Success() {
	ctx := context.Background()
	
	// Create test game sessions
	gameSessions := []*models.GameSession{
		{
			ID:          "test-game-7",
			Player1:     "target-player",
			Player2:     "other-player",
			CurrentTurn: models.PlayerColorRed,
			Status:      models.StatusCompleted,
			Board:       models.NewBoard(),
		},
		{
			ID:          "test-game-8",
			Player1:     "another-player",
			Player2:     "target-player",
			CurrentTurn: models.PlayerColorRed,
			Status:      models.StatusCompleted,
			Board:       models.NewBoard(),
		},
	}

	for _, gameSession := range gameSessions {
		err := suite.db.Create(gameSession).Error
		suite.Require().NoError(err)
	}

	// Get games by player
	playerGames, err := suite.repo.GetGamesByPlayer(ctx, "target-player")
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(playerGames), 2)

	// Verify all games include the target player
	for _, game := range playerGames {
		assert.True(suite.T(), 
			game.Player1 == "target-player" || game.Player2 == "target-player",
			"Game should include target player")
	}
}

func (suite *GameSessionRepositoryTestSuite) TestGetGameHistory_Success() {
	ctx := context.Background()
	
	// Create completed game session
	now := time.Now()
	gameSession := &models.GameSession{
		ID:          "test-game-9",
		Player1:     "player1",
		Player2:     "player2",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusCompleted,
		Board:       models.NewBoard(),
		EndTime:     &now,
	}
	err := suite.db.Create(gameSession).Error
	suite.Require().NoError(err)

	// Get game history
	history, err := suite.repo.GetGameHistory(ctx, 10, 0)
	assert.NoError(suite.T(), err)
	
	// Should contain at least the completed game
	completedCount := 0
	for _, game := range history {
		if game.Status == models.StatusCompleted {
			completedCount++
		}
	}
	assert.GreaterOrEqual(suite.T(), completedCount, 1)
}

func TestGameSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GameSessionRepositoryTestSuite))
}