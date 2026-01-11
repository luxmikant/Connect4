//go:build integration
// +build integration

package handlers_test

import (
	"bytes"
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

type GameHandlerIntegrationTestSuite struct {
	suite.Suite
	router      *gin.Engine
	gameHandler *handlers.GameHandler
	repoManager *repositories.Manager
}

func (suite *GameHandlerIntegrationTestSuite) SetupSuite() {
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
	suite.gameHandler = handlers.NewGameHandler(gameService)
	leaderboardHandler := handlers.NewLeaderboardHandler(repoManager.PlayerStats)

	// Setup router
	suite.router = gin.New()
	testConfig := &config.Config{
		Server: config.ServerConfig{
			CORSOrigins: []string{"*"},
		},
	}
	routes.SetupRoutes(suite.router, testConfig, suite.gameHandler, leaderboardHandler)
}

func (suite *GameHandlerIntegrationTestSuite) TestCreateGame() {
	// Test successful game creation
	createReq := handlers.CreateGameRequest{
		Player1: "alice",
		Player2: "bob",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.GameSession
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "alice", response.Player1)
	assert.Equal(suite.T(), "bob", response.Player2)
	assert.Equal(suite.T(), models.StatusInProgress, response.Status)
	assert.Equal(suite.T(), models.PlayerColorRed, response.CurrentTurn)
}

func (suite *GameHandlerIntegrationTestSuite) TestCreateGameValidation() {
	// Test validation errors
	testCases := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "same players",
			request: handlers.CreateGameRequest{
				Player1: "alice",
				Player2: "alice",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Players must have different usernames",
		},
		{
			name: "empty player1",
			request: handlers.CreateGameRequest{
				Player1: "",
				Player2: "bob",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "short username",
			request: handlers.CreateGameRequest{
				Player1: "ab",
				Player2: "bob",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req := httptest.NewRequest("POST", "/api/v1/games", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var errorResp handlers.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			assert.NoError(t, err)
			assert.Contains(t, errorResp.Error, tc.expectedError)
		})
	}
}

func (suite *GameHandlerIntegrationTestSuite) TestGetGameState() {
	// Create a game first
	session, err := suite.createTestGame("alice", "bob")
	suite.Require().NoError(err)

	// Test getting game state
	req := httptest.NewRequest("GET", "/api/v1/games/"+session.ID, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.GameSession
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, response.ID)
	assert.Equal(suite.T(), "alice", response.Player1)
	assert.Equal(suite.T(), "bob", response.Player2)
}

func (suite *GameHandlerIntegrationTestSuite) TestGetGameStateNotFound() {
	req := httptest.NewRequest("GET", "/api/v1/games/nonexistent", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var errorResp handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Game not found", errorResp.Error)
}

func (suite *GameHandlerIntegrationTestSuite) TestMakeMove() {
	// Create a game first
	session, err := suite.createTestGame("alice", "bob")
	suite.Require().NoError(err)

	// Test making a valid move
	moveReq := handlers.MakeMoveRequest{
		Column: 3,
		Player: "alice", // Player1 (red) goes first
	}

	body, _ := json.Marshal(moveReq)
	req := httptest.NewRequest("POST", "/api/v1/games/"+session.ID+"/moves", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.GameSession
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.PlayerColorYellow, response.CurrentTurn) // Turn switched to bob
	assert.Equal(suite.T(), models.PlayerColorRed, response.Board.Grid[0][3]) // Move was made
}

func (suite *GameHandlerIntegrationTestSuite) TestMakeMoveValidation() {
	// Create a game first
	session, err := suite.createTestGame("alice", "bob")
	suite.Require().NoError(err)

	testCases := []struct {
		name           string
		request        handlers.MakeMoveRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "wrong player turn",
			request: handlers.MakeMoveRequest{
				Column: 3,
				Player: "bob", // Should be alice's turn
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "Not your turn",
		},
		{
			name: "invalid column",
			request: handlers.MakeMoveRequest{
				Column: 7, // Out of bounds
				Player: "alice",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
		{
			name: "negative column",
			request: handlers.MakeMoveRequest{
				Column: -1,
				Player: "alice",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Validation failed",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req := httptest.NewRequest("POST", "/api/v1/games/"+session.ID+"/moves", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var errorResp handlers.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			assert.NoError(t, err)
			assert.Contains(t, errorResp.Error, tc.expectedError)
		})
	}
}

func (suite *GameHandlerIntegrationTestSuite) TestCompleteGameFlow() {
	// Create a game
	session, err := suite.createTestGame("alice", "bob")
	suite.Require().NoError(err)

	// Play a sequence of moves to create a win condition
	moves := []struct {
		player string
		column int
	}{
		{"alice", 0}, // Red
		{"bob", 1},   // Yellow
		{"alice", 0}, // Red
		{"bob", 1},   // Yellow
		{"alice", 0}, // Red
		{"bob", 1},   // Yellow
		{"alice", 0}, // Red - This should win (4 in a column)
	}

	var currentSession *models.GameSession
	for i, move := range moves {
		moveReq := handlers.MakeMoveRequest{
			Column: move.column,
			Player: move.player,
		}

		body, _ := json.Marshal(moveReq)
		req := httptest.NewRequest("POST", "/api/v1/games/"+session.ID+"/moves", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code, "Move %d failed", i+1)

		err = json.Unmarshal(w.Body.Bytes(), &currentSession)
		assert.NoError(suite.T(), err)

		// Check if this is the winning move
		if i == len(moves)-1 {
			// Alice should have won
			assert.Equal(suite.T(), models.StatusCompleted, currentSession.Status)
			assert.NotNil(suite.T(), currentSession.Winner)
			assert.Equal(suite.T(), models.PlayerColorRed, *currentSession.Winner)
		} else {
			// Game should still be in progress
			assert.Equal(suite.T(), models.StatusInProgress, currentSession.Status)
		}
	}
}

func (suite *GameHandlerIntegrationTestSuite) TestMakeMoveOnCompletedGame() {
	// Create a game and complete it
	session, err := suite.createTestGame("alice", "bob")
	suite.Require().NoError(err)

	// Complete the game by making alice win
	moves := []struct {
		player string
		column int
	}{
		{"alice", 0}, {"bob", 1}, {"alice", 0}, {"bob", 1},
		{"alice", 0}, {"bob", 1}, {"alice", 0}, // Alice wins
	}

	for _, move := range moves {
		moveReq := handlers.MakeMoveRequest{
			Column: move.column,
			Player: move.player,
		}

		body, _ := json.Marshal(moveReq)
		req := httptest.NewRequest("POST", "/api/v1/games/"+session.ID+"/moves", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}

	// Try to make another move on the completed game
	moveReq := handlers.MakeMoveRequest{
		Column: 2,
		Player: "bob",
	}

	body, _ := json.Marshal(moveReq)
	req := httptest.NewRequest("POST", "/api/v1/games/"+session.ID+"/moves", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var errorResp handlers.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Game is not active", errorResp.Error)
}

// Helper method to create a test game
func (suite *GameHandlerIntegrationTestSuite) createTestGame(player1, player2 string) (*models.GameSession, error) {
	createReq := handlers.CreateGameRequest{
		Player1: player1,
		Player2: player2,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		return nil, assert.AnError
	}

	var session models.GameSession
	err := json.Unmarshal(w.Body.Bytes(), &session)
	return &session, err
}

func TestGameHandlerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GameHandlerIntegrationTestSuite))
}