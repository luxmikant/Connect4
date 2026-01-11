package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"connect4-multiplayer/pkg/models"
)

// MockPlayerStatsRepository is a mock implementation of PlayerStatsRepository
type MockPlayerStatsRepository struct {
	mock.Mock
}

func (m *MockPlayerStatsRepository) Create(ctx context.Context, stats *models.PlayerStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) GetByID(ctx context.Context, id string) (*models.PlayerStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) GetByUsername(ctx context.Context, username string) (*models.PlayerStats, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) Update(ctx context.Context, stats *models.PlayerStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) UpdateGameStats(ctx context.Context, username string, won bool, gameDuration int) error {
	args := m.Called(ctx, username, won, gameDuration)
	return args.Error(0)
}

// PlayerStatsServiceTestSuite defines the test suite
type PlayerStatsServiceTestSuite struct {
	suite.Suite
	service  PlayerStatsService
	mockRepo *MockPlayerStatsRepository
	ctx      context.Context
}

func (suite *PlayerStatsServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockPlayerStatsRepository)
	suite.service = NewPlayerStatsService(suite.mockRepo, &ServiceConfig{
		StatsCacheTTL:       1 * time.Minute,
		LeaderboardCacheTTL: 30 * time.Second,
	})
	suite.ctx = context.Background()
}

func (suite *PlayerStatsServiceTestSuite) TestGetPlayerStats_Success() {
	expectedStats := &models.PlayerStats{
		ID:          "stats-1",
		Username:    "player1",
		GamesPlayed: 10,
		GamesWon:    6,
		WinRate:     0.6,
	}

	suite.mockRepo.On("GetByUsername", suite.ctx, "player1").Return(expectedStats, nil)

	stats, err := suite.service.GetPlayerStats(suite.ctx, "player1")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedStats, stats)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *PlayerStatsServiceTestSuite) TestGetPlayerStats_EmptyUsername() {
	stats, err := suite.service.GetPlayerStats(suite.ctx, "")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), stats)
	assert.Contains(suite.T(), err.Error(), "username cannot be empty")
}

func (suite *PlayerStatsServiceTestSuite) TestGetPlayerStats_NotFound() {
	suite.mockRepo.On("GetByUsername", suite.ctx, "unknown").Return(nil, models.ErrPlayerNotFound)

	stats, err := suite.service.GetPlayerStats(suite.ctx, "unknown")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), stats)
	assert.Equal(suite.T(), models.ErrPlayerNotFound, err)
}


func (suite *PlayerStatsServiceTestSuite) TestGetLeaderboard_Success() {
	expectedLeaderboard := []*models.PlayerStats{
		{Username: "player1", GamesWon: 10, WinRate: 0.8},
		{Username: "player2", GamesWon: 8, WinRate: 0.7},
		{Username: "player3", GamesWon: 5, WinRate: 0.5},
	}

	suite.mockRepo.On("GetLeaderboard", suite.ctx, 10).Return(expectedLeaderboard, nil)

	leaderboard, err := suite.service.GetLeaderboard(suite.ctx, 10)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedLeaderboard, leaderboard)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *PlayerStatsServiceTestSuite) TestGetLeaderboard_DefaultLimit() {
	expectedLeaderboard := []*models.PlayerStats{
		{Username: "player1", GamesWon: 10},
	}

	suite.mockRepo.On("GetLeaderboard", suite.ctx, 10).Return(expectedLeaderboard, nil)

	leaderboard, err := suite.service.GetLeaderboard(suite.ctx, 0)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), leaderboard)
}

func (suite *PlayerStatsServiceTestSuite) TestGetLeaderboard_MaxLimit() {
	expectedLeaderboard := []*models.PlayerStats{}

	suite.mockRepo.On("GetLeaderboard", suite.ctx, 100).Return(expectedLeaderboard, nil)

	// Request more than max limit
	leaderboard, err := suite.service.GetLeaderboard(suite.ctx, 200)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), leaderboard)
}

func (suite *PlayerStatsServiceTestSuite) TestRecordGameResult_Success() {
	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player1", true, 120).Return(nil)
	suite.mockRepo.On("GetLeaderboard", mock.Anything, 10).Return([]*models.PlayerStats{}, nil).Maybe()

	err := suite.service.RecordGameResult(suite.ctx, "player1", true, 120)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player1", true, 120)
}

func (suite *PlayerStatsServiceTestSuite) TestRecordGameResult_EmptyUsername() {
	err := suite.service.RecordGameResult(suite.ctx, "", true, 120)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "username cannot be empty")
}

func (suite *PlayerStatsServiceTestSuite) TestRecordGameCompletion_Player1Wins() {
	winner := models.PlayerColorRed

	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player1", true, 180).Return(nil)
	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player2", false, 180).Return(nil)
	suite.mockRepo.On("GetLeaderboard", mock.Anything, 10).Return([]*models.PlayerStats{}, nil).Maybe()

	err := suite.service.RecordGameCompletion(suite.ctx, "player1", "player2", &winner, 180)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player1", true, 180)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player2", false, 180)
}

func (suite *PlayerStatsServiceTestSuite) TestRecordGameCompletion_Player2Wins() {
	winner := models.PlayerColorYellow

	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player1", false, 180).Return(nil)
	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player2", true, 180).Return(nil)
	suite.mockRepo.On("GetLeaderboard", mock.Anything, 10).Return([]*models.PlayerStats{}, nil).Maybe()

	err := suite.service.RecordGameCompletion(suite.ctx, "player1", "player2", &winner, 180)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player1", false, 180)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player2", true, 180)
}

func (suite *PlayerStatsServiceTestSuite) TestRecordGameCompletion_Draw() {
	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player1", false, 200).Return(nil)
	suite.mockRepo.On("UpdateGameStats", suite.ctx, "player2", false, 200).Return(nil)
	suite.mockRepo.On("GetLeaderboard", mock.Anything, 10).Return([]*models.PlayerStats{}, nil).Maybe()

	err := suite.service.RecordGameCompletion(suite.ctx, "player1", "player2", nil, 200)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player1", false, 200)
	suite.mockRepo.AssertCalled(suite.T(), "UpdateGameStats", suite.ctx, "player2", false, 200)
}

func (suite *PlayerStatsServiceTestSuite) TestCreatePlayerStats_Success() {
	suite.mockRepo.On("Create", suite.ctx, mock.AnythingOfType("*models.PlayerStats")).Return(nil)

	stats, err := suite.service.CreatePlayerStats(suite.ctx, "newplayer")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), "newplayer", stats.Username)
	assert.Equal(suite.T(), 0, stats.GamesPlayed)
	assert.Equal(suite.T(), 0, stats.GamesWon)
}

func (suite *PlayerStatsServiceTestSuite) TestGetOrCreatePlayerStats_Existing() {
	existingStats := &models.PlayerStats{
		Username:    "existing",
		GamesPlayed: 5,
		GamesWon:    3,
	}

	suite.mockRepo.On("GetByUsername", suite.ctx, "existing").Return(existingStats, nil)

	stats, err := suite.service.GetOrCreatePlayerStats(suite.ctx, "existing")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), existingStats, stats)
}

func (suite *PlayerStatsServiceTestSuite) TestGetOrCreatePlayerStats_New() {
	suite.mockRepo.On("GetByUsername", suite.ctx, "newuser").Return(nil, models.ErrPlayerNotFound)
	suite.mockRepo.On("Create", suite.ctx, mock.AnythingOfType("*models.PlayerStats")).Return(nil)

	stats, err := suite.service.GetOrCreatePlayerStats(suite.ctx, "newuser")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), "newuser", stats.Username)
}

func (suite *PlayerStatsServiceTestSuite) TestLeaderboardSubscription() {
	var receivedLeaderboard []*models.PlayerStats
	var wg sync.WaitGroup
	wg.Add(1)

	callback := func(leaderboard []*models.PlayerStats) {
		receivedLeaderboard = leaderboard
		wg.Done()
	}

	subID := suite.service.SubscribeToLeaderboardUpdates(callback)
	assert.NotEmpty(suite.T(), subID)

	expectedLeaderboard := []*models.PlayerStats{
		{Username: "player1", GamesWon: 10},
	}
	suite.mockRepo.On("GetLeaderboard", mock.Anything, 10).Return(expectedLeaderboard, nil)

	suite.service.NotifyLeaderboardUpdate(suite.ctx)

	// Wait for callback with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(suite.T(), expectedLeaderboard, receivedLeaderboard)
	case <-time.After(2 * time.Second):
		suite.T().Fatal("timeout waiting for leaderboard update callback")
	}

	// Unsubscribe
	suite.service.UnsubscribeFromLeaderboardUpdates(subID)
}

func (suite *PlayerStatsServiceTestSuite) TestCacheStats() {
	stats := &models.PlayerStats{
		Username:    "cached_player",
		GamesPlayed: 10,
		GamesWon:    5,
	}

	suite.mockRepo.On("GetByUsername", suite.ctx, "cached_player").Return(stats, nil).Once()

	// First call should hit the database
	result1, err := suite.service.GetPlayerStats(suite.ctx, "cached_player")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), stats, result1)

	// Second call should use cache (no additional mock call)
	result2, err := suite.service.GetPlayerStats(suite.ctx, "cached_player")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), stats, result2)

	// Verify only one database call was made
	suite.mockRepo.AssertNumberOfCalls(suite.T(), "GetByUsername", 1)
}

func (suite *PlayerStatsServiceTestSuite) TestInvalidateCache() {
	stats := &models.PlayerStats{
		Username:    "player_to_invalidate",
		GamesPlayed: 10,
	}

	suite.mockRepo.On("GetByUsername", suite.ctx, "player_to_invalidate").Return(stats, nil)

	// First call
	_, _ = suite.service.GetPlayerStats(suite.ctx, "player_to_invalidate")

	// Invalidate cache
	suite.service.InvalidateCache("player_to_invalidate")

	// Second call should hit database again
	_, _ = suite.service.GetPlayerStats(suite.ctx, "player_to_invalidate")

	// Verify two database calls were made
	suite.mockRepo.AssertNumberOfCalls(suite.T(), "GetByUsername", 2)
}

func (suite *PlayerStatsServiceTestSuite) TestGetCacheStats() {
	cacheStats := suite.service.GetCacheStats()

	assert.Contains(suite.T(), cacheStats, "stats_cache_size")
	assert.Contains(suite.T(), cacheStats, "leaderboard_cached")
	assert.Contains(suite.T(), cacheStats, "subscription_count")
}

func TestPlayerStatsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PlayerStatsServiceTestSuite))
}
