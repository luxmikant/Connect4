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

type PlayerStatsRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo repositories.PlayerStatsRepository
}

func (suite *PlayerStatsRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.PlayerStats{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repositories.NewPlayerStatsRepository(db)
}

func (suite *PlayerStatsRepositoryTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *PlayerStatsRepositoryTestSuite) TestCreate_Success() {
	ctx := context.Background()
	stats := &models.PlayerStats{
		ID:          "test-stats-1",
		Username:    "testuser",
		GamesPlayed: 10,
		GamesWon:    7,
		WinRate:     0.7,
		AvgGameTime: 300,
		LastPlayed:  time.Now(),
	}

	err := suite.repo.Create(ctx, stats)
	assert.NoError(suite.T(), err)

	// Verify stats were created
	var retrieved models.PlayerStats
	err = suite.db.First(&retrieved, "id = ?", stats.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), stats.Username, retrieved.Username)
	assert.Equal(suite.T(), stats.GamesPlayed, retrieved.GamesPlayed)
	assert.Equal(suite.T(), stats.GamesWon, retrieved.GamesWon)
}

func (suite *PlayerStatsRepositoryTestSuite) TestGetByUsername_Success() {
	ctx := context.Background()
	
	// Create test stats
	stats := &models.PlayerStats{
		ID:          "test-stats-2",
		Username:    "testuser2",
		GamesPlayed: 5,
		GamesWon:    3,
		WinRate:     0.6,
		AvgGameTime: 250,
		LastPlayed:  time.Now(),
	}
	err := suite.db.Create(stats).Error
	suite.Require().NoError(err)

	// Retrieve stats by username
	retrieved, err := suite.repo.GetByUsername(ctx, stats.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), stats.ID, retrieved.ID)
	assert.Equal(suite.T(), stats.Username, retrieved.Username)
	assert.Equal(suite.T(), stats.GamesPlayed, retrieved.GamesPlayed)
}

func (suite *PlayerStatsRepositoryTestSuite) TestGetByUsername_NotFound() {
	ctx := context.Background()
	
	retrieved, err := suite.repo.GetByUsername(ctx, "non-existent-user")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrieved)
	assert.Equal(suite.T(), models.ErrPlayerNotFound, err)
}

func (suite *PlayerStatsRepositoryTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	
	// Create test stats
	stats := &models.PlayerStats{
		ID:          "test-stats-3",
		Username:    "testuser3",
		GamesPlayed: 8,
		GamesWon:    4,
		WinRate:     0.5,
		AvgGameTime: 280,
		LastPlayed:  time.Now(),
	}
	err := suite.db.Create(stats).Error
	suite.Require().NoError(err)

	// Update stats
	stats.GamesPlayed = 10
	stats.GamesWon = 6
	stats.CalculateWinRate()
	err = suite.repo.Update(ctx, stats)
	assert.NoError(suite.T(), err)

	// Verify update
	var retrieved models.PlayerStats
	err = suite.db.First(&retrieved, "id = ?", stats.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 10, retrieved.GamesPlayed)
	assert.Equal(suite.T(), 6, retrieved.GamesWon)
	assert.Equal(suite.T(), 0.6, retrieved.WinRate)
}

func (suite *PlayerStatsRepositoryTestSuite) TestGetLeaderboard_Success() {
	ctx := context.Background()
	
	// Create multiple test stats with different win counts
	statsData := []*models.PlayerStats{
		{
			ID:          "test-stats-4",
			Username:    "topplayer",
			GamesPlayed: 20,
			GamesWon:    18,
			WinRate:     0.9,
			AvgGameTime: 300,
			LastPlayed:  time.Now(),
		},
		{
			ID:          "test-stats-5",
			Username:    "midplayer",
			GamesPlayed: 15,
			GamesWon:    10,
			WinRate:     0.67,
			AvgGameTime: 280,
			LastPlayed:  time.Now(),
		},
		{
			ID:          "test-stats-6",
			Username:    "newplayer",
			GamesPlayed: 5,
			GamesWon:    2,
			WinRate:     0.4,
			AvgGameTime: 320,
			LastPlayed:  time.Now(),
		},
	}

	for _, stats := range statsData {
		err := suite.db.Create(stats).Error
		suite.Require().NoError(err)
	}

	// Get leaderboard
	leaderboard, err := suite.repo.GetLeaderboard(ctx, 10)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(leaderboard), 3)

	// Verify ordering (should be sorted by wins descending)
	if len(leaderboard) >= 2 {
		assert.GreaterOrEqual(suite.T(), leaderboard[0].GamesWon, leaderboard[1].GamesWon)
	}
}

func (suite *PlayerStatsRepositoryTestSuite) TestUpdateGameStats_NewPlayer() {
	ctx := context.Background()
	
	// Update stats for a new player (should create new record)
	err := suite.repo.UpdateGameStats(ctx, "newplayer", true, 300)
	assert.NoError(suite.T(), err)

	// Verify new stats were created
	var stats models.PlayerStats
	err = suite.db.First(&stats, "username = ?", "newplayer").Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, stats.GamesPlayed)
	assert.Equal(suite.T(), 1, stats.GamesWon)
	assert.Equal(suite.T(), 1.0, stats.WinRate)
	assert.Equal(suite.T(), 300, stats.AvgGameTime)
}

func (suite *PlayerStatsRepositoryTestSuite) TestUpdateGameStats_ExistingPlayer() {
	ctx := context.Background()
	
	// Create existing player stats
	stats := &models.PlayerStats{
		ID:          "test-stats-7",
		Username:    "existingplayer",
		GamesPlayed: 5,
		GamesWon:    3,
		WinRate:     0.6,
		AvgGameTime: 250,
		LastPlayed:  time.Now().Add(-time.Hour),
	}
	err := suite.db.Create(stats).Error
	suite.Require().NoError(err)

	// Update stats with a loss
	err = suite.repo.UpdateGameStats(ctx, "existingplayer", false, 400)
	assert.NoError(suite.T(), err)

	// Verify stats were updated
	var retrieved models.PlayerStats
	err = suite.db.First(&retrieved, "username = ?", "existingplayer").Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 6, retrieved.GamesPlayed)
	assert.Equal(suite.T(), 3, retrieved.GamesWon)
	assert.Equal(suite.T(), 0.5, retrieved.WinRate)
	// Average game time should be updated: (250*5 + 400) / 6 = 275
	assert.Equal(suite.T(), 275, retrieved.AvgGameTime)
}

func (suite *PlayerStatsRepositoryTestSuite) TestUpdateGameStats_EmptyUsername() {
	ctx := context.Background()
	
	err := suite.repo.UpdateGameStats(ctx, "", true, 300)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "username cannot be empty")
}

func TestPlayerStatsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PlayerStatsRepositoryTestSuite))
}