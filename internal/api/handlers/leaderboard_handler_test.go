//go:build integration
// +build integration

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"connect4-multiplayer/internal/api/handlers"
	"connect4-multiplayer/internal/api/routes"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/pkg/models"
)

type LeaderboardHandlerIntegrationTestSuite struct {
	suite.Suite
	router             *gin.Engine
	leaderboardHandler *handlers.LeaderboardHandler
	repoManager        *repositories.Manager
}

func (suite *LeaderboardHandlerIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	_, repoManager, err := initializeTestDatabase()
	suite.Require().NoError(err)
	suite.repoManager = repoManager

	// Initialize services
	gameService := game.NewGameService(
		repoManager.GameSession,
		repoManager.PlayerStats,
		repoManager.Move,
		repoManager.GameEvent,
		game.DefaultServiceConfig(),
	)

	// Initialize handlers
	gameHandler := handlers.NewGameHandler(gameService)
	suite.leaderboardHandler = handlers.NewLeaderboardHandler(repoManager.PlayerStats)

	// Setup router
	suite.router = gin.New()
	testConfig := &config.Config{
		Server: config.ServerConfig{
			CORSOrigins: []string{"*"},
		},
	}
	routes.SetupRoutes(suite.router, testConfig, gameHandler, suite.leaderboardHandler)
}

func (suite *LeaderboardHandlerIntegrationTestSuite) SetupTest() {
	// Create test player statistics
	suite.createTestPlayerStats()
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetLeaderboard() {
	// Test getting leaderboard with default limit
	req := httptest.NewRequest("GET", "/api/v1/leaderboard", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []*models.PlayerStats
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), len(response), 10) // Default limit is 10

	// Verify ordering (should be sorted by wins descending)
	if len(response) > 1 {
		for i := 0; i < len(response)-1; i++ {
			assert.GreaterOrEqual(suite.T(), response[i].GamesWon, response[i+1].GamesWon)
		}
	}
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetLeaderboardWithLimit() {
	// Test getting leaderboard with custom limit
	req := httptest.NewRequest("GET", "/api/v1/leaderboard?limit=5", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []*models.PlayerStats
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), len(response), 5)
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetLeaderboardWithInvalidLimit() {
	testCases := []struct {
		name  string
		limit string
	}{
		{"negative limit", "-1"},
		{"zero limit", "0"},
		{"invalid string", "abc"},
		{"excessive limit", "200"}, // Should be capped at 100
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/leaderboard?limit="+tc.limit, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response []*models.PlayerStats
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Should use default limit (10) for invalid values, or cap at 100
			if tc.limit == "200" {
				assert.LessOrEqual(t, len(response), 100)
			} else {
				assert.LessOrEqual(t, len(response), 10)
			}
		})
	}
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetPlayerStats() {
	// Test getting existing player stats
	req := httptest.NewRequest("GET", "/api/v1/players/alice/stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.PlayerStats
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "alice", response.Username)
	assert.Greater(suite.T(), response.GamesPlayed, 0)
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetPlayerStatsNotFound() {
	// Test getting non-existent player stats
	req := httptest.NewRequest("GET", "/api/v1/players/nonexistent/stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var errorResp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Player not found", errorResp.Error)
}

func (suite *LeaderboardHandlerIntegrationTestSuite) TestGetPlayerStatsEmptyUsername() {
	// Test getting player stats with empty username
	req := httptest.NewRequest("GET", "/api/v1/players//stats", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// This should result in a 404 because the route won't match
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Helper method to create test player statistics
func (suite *LeaderboardHandlerIntegrationTestSuite) createTestPlayerStats() {
	ctx := suite.T().Context()

	// Create test players with different stats
	players := []struct {
		username    string
		gamesPlayed int
		gamesWon    int
	}{
		{"alice", 10, 8},   // 80% win rate
		{"bob", 15, 9},     // 60% win rate
		{"charlie", 8, 6},  // 75% win rate
		{"diana", 12, 4},   // 33% win rate
		{"eve", 20, 15},    // 75% win rate
	}

	for _, player := range players {
		stats := &models.PlayerStats{
			Username:    player.username,
			GamesPlayed: player.gamesPlayed,
			GamesWon:    player.gamesWon,
		}
		stats.CalculateWinRate()

		err := suite.repoManager.PlayerStats.Create(ctx, stats)
		suite.Require().NoError(err)
	}
}

func TestLeaderboardHandlerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LeaderboardHandlerIntegrationTestSuite))
}