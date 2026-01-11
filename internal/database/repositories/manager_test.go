package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

type ManagerTestSuite struct {
	suite.Suite
	db      *gorm.DB
	manager *repositories.Manager
}

func (suite *ManagerTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Player{},
		&models.GameSession{},
		&models.Move{},
		&models.PlayerStats{},
		&models.GameEvent{},
	)
	suite.Require().NoError(err)

	suite.db = db
	suite.manager = repositories.NewManager(db)
}

func (suite *ManagerTestSuite) TearDownTest() {
	if suite.manager != nil {
		suite.manager.Close()
	}
}

func (suite *ManagerTestSuite) TestNewManager_Success() {
	assert.NotNil(suite.T(), suite.manager)
	assert.NotNil(suite.T(), suite.manager.Player)
	assert.NotNil(suite.T(), suite.manager.GameSession)
	assert.NotNil(suite.T(), suite.manager.PlayerStats)
	assert.NotNil(suite.T(), suite.manager.Move)
	assert.NotNil(suite.T(), suite.manager.GameEvent)
}

func (suite *ManagerTestSuite) TestHealthCheck_Success() {
	ctx := context.Background()
	
	err := suite.manager.HealthCheck(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *ManagerTestSuite) TestGetConnectionStats_Success() {
	stats, err := suite.manager.GetConnectionStats()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	
	// Verify expected fields are present
	expectedFields := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
		"max_idle_closed",
		"max_idle_time_closed",
		"max_lifetime_closed",
	}
	
	for _, field := range expectedFields {
		assert.Contains(suite.T(), stats, field)
	}
}

func (suite *ManagerTestSuite) TestWithTransaction_Success() {
	ctx := context.Background()
	
	// Test successful transaction
	err := suite.manager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Create a player within transaction
		player := &models.Player{
			ID:       "test-tx-player",
			Username: "txuser",
		}
		return tx.Create(player).Error
	})
	
	assert.NoError(suite.T(), err)
	
	// Verify player was created
	var player models.Player
	err = suite.db.First(&player, "id = ?", "test-tx-player").Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "txuser", player.Username)
}

func (suite *ManagerTestSuite) TestWithTransaction_Rollback() {
	ctx := context.Background()
	
	// Test transaction rollback on error
	err := suite.manager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Create a player within transaction
		player := &models.Player{
			ID:       "test-rollback-player",
			Username: "rollbackuser",
		}
		if err := tx.Create(player).Error; err != nil {
			return err
		}
		
		// Force an error to trigger rollback
		return assert.AnError
	})
	
	assert.Error(suite.T(), err)
	
	// Verify player was NOT created (transaction rolled back)
	var player models.Player
	err = suite.db.First(&player, "id = ?", "test-rollback-player").Error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *ManagerTestSuite) TestBeginTransaction_Success() {
	ctx := context.Background()
	
	tx, err := suite.manager.BeginTransaction(ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tx)
	
	// Test that we can use the transaction
	player := &models.Player{
		ID:       "test-manual-tx-player",
		Username: "manualtxuser",
	}
	err = tx.Create(player).Error
	assert.NoError(suite.T(), err)
	
	// Commit the transaction
	err = tx.Commit().Error
	assert.NoError(suite.T(), err)
	
	// Verify player was created
	var retrieved models.Player
	err = suite.db.First(&retrieved, "id = ?", "test-manual-tx-player").Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "manualtxuser", retrieved.Username)
}

func (suite *ManagerTestSuite) TestRepositoryIntegration() {
	ctx := context.Background()
	
	// Test that all repositories work together
	
	// 1. Create a player
	player := &models.Player{
		ID:       "integration-player",
		Username: "integrationuser",
	}
	err := suite.manager.Player.Create(ctx, player)
	assert.NoError(suite.T(), err)
	
	// 2. Create player stats
	stats := &models.PlayerStats{
		ID:          "integration-stats",
		Username:    "integrationuser",
		GamesPlayed: 0,
		GamesWon:    0,
		WinRate:     0.0,
	}
	err = suite.manager.PlayerStats.Create(ctx, stats)
	assert.NoError(suite.T(), err)
	
	// 3. Create a game session
	gameSession := &models.GameSession{
		ID:          "integration-game",
		Player1:     "integrationuser",
		Player2:     "opponent",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		Board:       models.NewBoard(),
	}
	err = suite.manager.GameSession.Create(ctx, gameSession)
	assert.NoError(suite.T(), err)
	
	// 4. Create a move
	move := &models.Move{
		ID:     "integration-move",
		GameID: "integration-game",
		Player: models.PlayerColorRed,
		Column: 3,
		Row:    0,
	}
	err = suite.manager.Move.Create(ctx, move)
	assert.NoError(suite.T(), err)
	
	// 5. Create a game event
	event := &models.GameEvent{
		ID:        "integration-event",
		EventType: models.EventMoveMade,
		GameID:    "integration-game",
		PlayerID:  "integrationuser",
	}
	err = suite.manager.GameEvent.Create(ctx, event)
	assert.NoError(suite.T(), err)
	
	// Verify all data was created correctly
	retrievedPlayer, err := suite.manager.Player.GetByID(ctx, "integration-player")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "integrationuser", retrievedPlayer.Username)
	
	retrievedGame, err := suite.manager.GameSession.GetByID(ctx, "integration-game")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "integrationuser", retrievedGame.Player1)
	
	retrievedMove, err := suite.manager.Move.GetByID(ctx, "integration-move")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "integration-game", retrievedMove.GameID)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}