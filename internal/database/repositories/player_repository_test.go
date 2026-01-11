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

type PlayerRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repositories.PlayerRepository
}

func (suite *PlayerRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Player{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repositories.NewPlayerRepository(db)
}

func (suite *PlayerRepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *PlayerRepositoryTestSuite) TestCreate_Success() {
	ctx := context.Background()
	player := &models.Player{
		ID:       "test-player-1",
		Username: "testuser",
	}

	err := suite.repo.Create(ctx, player)
	assert.NoError(suite.T(), err)

	// Verify player was created
	var retrieved models.Player
	err = suite.db.First(&retrieved, "id = ?", player.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), player.Username, retrieved.Username)
}

func (suite *PlayerRepositoryTestSuite) TestCreate_NilPlayer() {
	ctx := context.Background()
	err := suite.repo.Create(ctx, nil)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "player cannot be nil")
}

func (suite *PlayerRepositoryTestSuite) TestGetByID_Success() {
	ctx := context.Background()
	
	// Create test player
	player := &models.Player{
		ID:       "test-player-2",
		Username: "testuser2",
	}
	err := suite.db.Create(player).Error
	suite.Require().NoError(err)

	// Retrieve player
	retrieved, err := suite.repo.GetByID(ctx, player.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), player.ID, retrieved.ID)
	assert.Equal(suite.T(), player.Username, retrieved.Username)
}

func (suite *PlayerRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	
	retrieved, err := suite.repo.GetByID(ctx, "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
	assert.Equal(suite.T(), models.ErrPlayerNotFound, err)
}

func (suite *PlayerRepositoryTestSuite) TestGetByID_EmptyID() {
	ctx := context.Background()
	
	retrieved, err := suite.repo.GetByID(ctx, "")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
	assert.Contains(suite.T(), err.Error(), "player ID cannot be empty")
}

func (suite *PlayerRepositoryTestSuite) TestGetByUsername_Success() {
	ctx := context.Background()
	
	// Create test player
	player := &models.Player{
		ID:       "test-player-3",
		Username: "testuser3",
	}
	err := suite.db.Create(player).Error
	suite.Require().NoError(err)

	// Retrieve player by username
	retrieved, err := suite.repo.GetByUsername(ctx, player.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), player.ID, retrieved.ID)
	assert.Equal(suite.T(), player.Username, retrieved.Username)
}

func (suite *PlayerRepositoryTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	
	// Create test player
	player := &models.Player{
		ID:       "test-player-4",
		Username: "testuser4",
	}
	err := suite.db.Create(player).Error
	suite.Require().NoError(err)

	// Update player
	player.Username = "updateduser4"
	err = suite.repo.Update(ctx, player)
	assert.NoError(suite.T(), err)

	// Verify update
	var retrieved models.Player
	err = suite.db.First(&retrieved, "id = ?", player.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updateduser4", retrieved.Username)
}

func (suite *PlayerRepositoryTestSuite) TestDelete_Success() {
	ctx := context.Background()
	
	// Create test player
	player := &models.Player{
		ID:       "test-player-5",
		Username: "testuser5",
	}
	err := suite.db.Create(player).Error
	suite.Require().NoError(err)

	// Delete player
	err = suite.repo.Delete(ctx, player.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	var retrieved models.Player
	err = suite.db.First(&retrieved, "id = ?", player.ID).Error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *PlayerRepositoryTestSuite) TestList_Success() {
	ctx := context.Background()
	
	// Create multiple test players
	players := []*models.Player{
		{ID: "test-player-6", Username: "testuser6"},
		{ID: "test-player-7", Username: "testuser7"},
		{ID: "test-player-8", Username: "testuser8"},
	}

	for _, player := range players {
		err := suite.db.Create(player).Error
		suite.Require().NoError(err)
	}

	// List players
	retrieved, err := suite.repo.List(ctx, 10, 0)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(retrieved), 3)
}

func (suite *PlayerRepositoryTestSuite) TestList_WithPagination() {
	ctx := context.Background()
	
	// Test pagination with limit
	retrieved, err := suite.repo.List(ctx, 2, 0)
	assert.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), len(retrieved), 2)
}

func TestPlayerRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PlayerRepositoryTestSuite))
}