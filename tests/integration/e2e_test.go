//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/api/handlers"
	"connect4-multiplayer/internal/api/middleware"
	"connect4-multiplayer/internal/api/routes"
	"connect4-multiplayer/internal/bot"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/internal/stats"
	wsService "connect4-multiplayer/internal/websocket"
	"connect4-multiplayer/pkg/models"
)

// E2ETestSuite provides end-to-end integration testing with cloud services
type E2ETestSuite struct {
	suite.Suite
	
	// Configuration
	config *config.Config
	
	// Database
	db         *gorm.DB
	repoManager *repositories.Manager
	
	// Services
	gameService        game.GameService
	matchmakingService matchmaking.MatchmakingService
	statsService       stats.PlayerStatsService
	analyticsService   *analytics.Service
	websocketService   *wsService.Service
	botService         bot.BotPlayerService
	
	// HTTP Server
	router *gin.Engine
	server *httptest.Server
	
	// Test context
	ctx    context.Context
	cancel context.CancelFunc
}

// SetupSuite initializes the test environment with cloud services
func (suite *E2ETestSuite) SetupSuite() {
	// Set up test context
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	
	// Load configuration
	cfg, err := config.Load()
	require.NoError(suite.T(), err, "Failed to load configuration")
	suite.config = cfg
	
	// Override config for testing
	suite.setupTestConfig()
	
	// Initialize database with cloud service (Supabase)
	suite.setupDatabase()
	
	// Initialize services
	suite.setupServices()
	
	// Initialize HTTP server
	suite.setupHTTPServer()
	
	// Start services
	suite.startServices()
}

// TearDownSuite cleans up the test environment
func (suite *E2ETestSuite) TearDownSuite() {
	// Stop services
	if suite.websocketService != nil {
		suite.websocketService.Stop()
	}
	
	if suite.analyticsService != nil {
		suite.analyticsService.Stop()
	}
	
	// Close server
	if suite.server != nil {
		suite.server.Close()
	}
	
	// Clean up database
	suite.cleanupDatabase()
	
	// Cancel context
	if suite.cancel != nil {
		suite.cancel()
	}
}

// setupTestConfig configures the test environment
func (suite *E2ETestSuite) setupTestConfig() {
	// Use test database if available, otherwise use configured database
	if testDB := os.Getenv("TEST_DATABASE_URL"); testDB != "" {
		suite.config.Database.URL = testDB
	}
	
	// Use test Kafka topic to avoid conflicts
	suite.config.Kafka.Topic = "test-game-events"
	suite.config.Kafka.ConsumerGroup = "test-analytics-service"
	
	// Set test environment
	suite.config.Environment = "test"
}

// setupDatabase initializes the database connection and runs migrations
func (suite *E2ETestSuite) setupDatabase() {
	// Connect to database (Supabase PostgreSQL)
	db, err := gorm.Open(postgres.Open(suite.config.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce noise in tests
	})
	require.NoError(suite.T(), err, "Failed to connect to database")
	
	suite.db = db
	
	// Run migrations
	migrator := database.NewMigrator(db)
	err = migrator.Up()
	require.NoError(suite.T(), err, "Failed to run database migrations")
	
	// Create repository manager
	suite.repoManager = repositories.NewManager(db)
	
	// Clean existing test data
	suite.cleanupTestData()
}

// setupServices initializes all application services
func (suite *E2ETestSuite) setupServices() {
	// Game service
	suite.gameService = game.NewGameService(
		suite.repoManager.GameSession,
		suite.repoManager.PlayerStats,
		suite.repoManager.Move,
		suite.repoManager.GameEvent,
		nil, // Use default config
	)
	
	// Bot service
	suite.botService = bot.NewBotPlayerService()
	
	// Stats service
	suite.statsService = stats.NewPlayerStatsService(
		suite.repoManager.PlayerStats,
		nil, // Use default config
	)
	
	// Matchmaking service
	matchmakingConfig := &matchmaking.ServiceConfig{
		MatchTimeout:  2 * time.Second, // Shorter timeout for tests
		MatchInterval: 100 * time.Millisecond,
		Logger:        slog.Default(),
	}
	suite.matchmakingService = matchmaking.NewMatchmakingService(suite.gameService, matchmakingConfig)
	
	// Analytics service (with Confluent Cloud Kafka)
	var analyticsService *analytics.Service
	var err error
	analyticsService, err = analytics.NewService(suite.config.Kafka, suite.db)
	require.NoError(suite.T(), err, "Failed to create analytics service")
	suite.analyticsService = analyticsService
	
	// WebSocket service
	suite.websocketService = wsService.NewService(suite.gameService, suite.matchmakingService)
}

// setupHTTPServer initializes the HTTP server with all routes
func (suite *E2ETestSuite) setupHTTPServer() {
	gin.SetMode(gin.TestMode)
	
	// Create router
	suite.router = gin.New()
	
	// Add middleware
	suite.router.Use(middleware.CORS(nil))
	suite.router.Use(middleware.Logging(nil))
	suite.router.Use(middleware.Recovery(nil))
	suite.router.Use(middleware.Validation(nil))
	
	// Create handlers
	gameHandler := handlers.NewGameHandler(suite.gameService)
	leaderboardHandler := handlers.NewLeaderboardHandler(suite.statsService)
	
	// Setup routes
	routes.SetupRoutes(suite.router, gameHandler, leaderboardHandler, suite.websocketService.GetWebSocketHandler())
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

// startServices starts all background services
func (suite *E2ETestSuite) startServices() {
	// Start WebSocket service
	err := suite.websocketService.Start(suite.ctx)
	require.NoError(suite.T(), err, "Failed to start WebSocket service")
	
	// Start analytics service
	err = suite.analyticsService.Start(suite.ctx)
	require.NoError(suite.T(), err, "Failed to start analytics service")
	
	// Allow services to initialize
	time.Sleep(500 * time.Millisecond)
}

// cleanupDatabase removes test data from database
func (suite *E2ETestSuite) cleanupDatabase() {
	if suite.db != nil {
		suite.cleanupTestData()
	}
}

// cleanupTestData removes all test data from database tables
func (suite *E2ETestSuite) cleanupTestData() {
	// Delete in reverse dependency order
	suite.db.Exec("DELETE FROM game_events WHERE game_id LIKE 'test-%'")
	suite.db.Exec("DELETE FROM moves WHERE game_id LIKE 'test-%'")
	suite.db.Exec("DELETE FROM game_sessions WHERE id LIKE 'test-%'")
	suite.db.Exec("DELETE FROM player_stats WHERE username LIKE 'test_%'")
	suite.db.Exec("DELETE FROM players WHERE username LIKE 'test_%'")
}

// TestCompleteGameFlow tests the entire game flow from matchmaking to completion
func (suite *E2ETestSuite) TestCompleteGameFlow() {
	// Test data
	player1 := "test_player1"
	player2 := "test_player2"
	
	// Step 1: Create players via API
	suite.createPlayer(player1)
	suite.createPlayer(player2)
	
	// Step 2: Test matchmaking via WebSocket
	conn1, conn2 := suite.setupWebSocketConnections(player1, player2)
	defer conn1.Close()
	defer conn2.Close()
	
	// Step 3: Join matchmaking queue
	gameID := suite.testMatchmaking(conn1, conn2, player1, player2)
	
	// Step 4: Play a complete game
	winner := suite.playCompleteGame(conn1, conn2, gameID, player1, player2)
	
	// Step 5: Verify game completion and statistics
	suite.verifyGameCompletion(gameID, winner)
	
	// Step 6: Verify analytics events were published
	suite.verifyAnalyticsEvents(gameID, player1, player2)
	
	// Step 7: Test leaderboard updates
	suite.verifyLeaderboardUpdates(player1, player2, winner)
}

// TestBotGameFlow tests player vs bot game flow
func (suite *E2ETestSuite) TestBotGameFlow() {
	player := "test_bot_player"
	
	// Create player
	suite.createPlayer(player)
	
	// Setup WebSocket connection
	conn := suite.setupSingleWebSocketConnection(player)
	defer conn.Close()
	
	// Join matchmaking (should timeout and create bot game)
	gameID := suite.testBotMatchmaking(conn, player)
	
	// Play against bot
	suite.playAgainstBot(conn, gameID, player)
	
	// Verify bot game completion
	suite.verifyBotGameCompletion(gameID)
}

// TestMultiClientScenarios tests multiple concurrent games
func (suite *E2ETestSuite) TestMultiClientScenarios() {
	numGames := 3
	players := make([]string, numGames*2)
	connections := make([]*websocket.Conn, numGames*2)
	
	// Create players and connections
	for i := 0; i < numGames*2; i++ {
		players[i] = fmt.Sprintf("test_multi_%d", i)
		suite.createPlayer(players[i])
		connections[i] = suite.setupSingleWebSocketConnection(players[i])
		defer connections[i].Close()
	}
	
	// Start multiple games concurrently
	gameIDs := make([]string, numGames)
	for i := 0; i < numGames; i++ {
		player1 := players[i*2]
		player2 := players[i*2+1]
		conn1 := connections[i*2]
		conn2 := connections[i*2+1]
		
		gameIDs[i] = suite.testMatchmaking(conn1, conn2, player1, player2)
	}
	
	// Verify all games are active
	for _, gameID := range gameIDs {
		session, err := suite.gameService.GetSession(suite.ctx, gameID)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), models.GameStatusInProgress, session.Status)
	}
	
	// Complete all games
	for i, gameID := range gameIDs {
		conn1 := connections[i*2]
		conn2 := connections[i*2+1]
		player1 := players[i*2]
		player2 := players[i*2+1]
		
		suite.playCompleteGame(conn1, conn2, gameID, player1, player2)
	}
}

// TestReconnectionScenarios tests WebSocket reconnection handling
func (suite *E2ETestSuite) TestReconnectionScenarios() {
	player1 := "test_reconnect1"
	player2 := "test_reconnect2"
	
	// Create players
	suite.createPlayer(player1)
	suite.createPlayer(player2)
	
	// Setup initial connections
	conn1, conn2 := suite.setupWebSocketConnections(player1, player2)
	
	// Start game
	gameID := suite.testMatchmaking(conn1, conn2, player1, player2)
	
	// Make a few moves
	suite.makeMove(conn1, gameID, 3) // Player 1 move
	suite.makeMove(conn2, gameID, 3) // Player 2 move
	
	// Simulate disconnection by closing connection
	conn1.Close()
	
	// Wait a moment
	time.Sleep(100 * time.Millisecond)
	
	// Reconnect player 1
	newConn1 := suite.setupSingleWebSocketConnection(player1)
	defer newConn1.Close()
	defer conn2.Close()
	
	// Send reconnection message
	reconnectMsg := map[string]interface{}{
		"type": "reconnect",
		"payload": map[string]interface{}{
			"gameId":   gameID,
			"username": player1,
		},
	}
	suite.sendWebSocketMessage(newConn1, reconnectMsg)
	
	// Verify reconnection successful
	response := suite.readWebSocketMessage(newConn1)
	assert.Equal(suite.T(), "game_state", response["type"])
	
	// Continue game
	suite.makeMove(newConn1, gameID, 4) // Player 1 move after reconnection
	
	// Verify game continues normally
	session, err := suite.gameService.GetSession(suite.ctx, gameID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.GameStatusInProgress, session.Status)
}

// TestKafkaIntegration tests Confluent Cloud Kafka integration
func (suite *E2ETestSuite) TestKafkaIntegration() {
	// Skip if Kafka credentials not available
	if suite.config.Kafka.APIKey == "" || suite.config.Kafka.APISecret == "" {
		suite.T().Skip("Kafka credentials not available for integration test")
	}
	
	player1 := "test_kafka1"
	player2 := "test_kafka2"
	
	// Create players
	suite.createPlayer(player1)
	suite.createPlayer(player2)
	
	// Setup connections and play game
	conn1, conn2 := suite.setupWebSocketConnections(player1, player2)
	defer conn1.Close()
	defer conn2.Close()
	
	gameID := suite.testMatchmaking(conn1, conn2, player1, player2)
	
	// Make some moves to generate events
	suite.makeMove(conn1, gameID, 3)
	suite.makeMove(conn2, gameID, 3)
	suite.makeMove(conn1, gameID, 4)
	
	// Wait for events to be processed
	time.Sleep(2 * time.Second)
	
	// Verify events were processed by analytics service
	// (This would check the database for processed analytics data)
	var eventCount int64
	err := suite.db.Model(&models.GameEvent{}).Where("game_id = ?", gameID).Count(&eventCount).Error
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), eventCount, int64(0), "Analytics events should be processed")
}

// Helper methods for test implementation

func (suite *E2ETestSuite) createPlayer(username string) {
	url := fmt.Sprintf("%s/api/v1/players", suite.server.URL)
	payload := fmt.Sprintf(`{"username": "%s"}`, username)
	
	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.True(suite.T(), resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
}

func (suite *E2ETestSuite) setupWebSocketConnections(player1, player2 string) (*websocket.Conn, *websocket.Conn) {
	conn1 := suite.setupSingleWebSocketConnection(player1)
	conn2 := suite.setupSingleWebSocketConnection(player2)
	return conn1, conn2
}

func (suite *E2ETestSuite) setupSingleWebSocketConnection(username string) *websocket.Conn {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(suite.server.URL, "http://", "ws://", 1) + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	
	// Send join message
	joinMsg := map[string]interface{}{
		"type": "join_game",
		"payload": map[string]interface{}{
			"username": username,
		},
	}
	suite.sendWebSocketMessage(conn, joinMsg)
	
	return conn
}

func (suite *E2ETestSuite) testMatchmaking(conn1, conn2 *websocket.Conn, player1, player2 string) string {
	// Both players should receive game_started message
	response1 := suite.readWebSocketMessage(conn1)
	response2 := suite.readWebSocketMessage(conn2)
	
	assert.Equal(suite.T(), "game_started", response1["type"])
	assert.Equal(suite.T(), "game_started", response2["type"])
	
	// Extract game ID
	payload1 := response1["payload"].(map[string]interface{})
	gameID := payload1["gameId"].(string)
	
	return gameID
}

func (suite *E2ETestSuite) testBotMatchmaking(conn *websocket.Conn, player string) string {
	// Wait for matchmaking timeout (should create bot game)
	time.Sleep(3 * time.Second)
	
	// Should receive game_started message with bot
	response := suite.readWebSocketMessage(conn)
	assert.Equal(suite.T(), "game_started", response["type"])
	
	payload := response["payload"].(map[string]interface{})
	gameID := payload["gameId"].(string)
	opponent := payload["opponent"].(string)
	
	assert.True(suite.T(), strings.Contains(opponent, "bot"), "Opponent should be a bot")
	
	return gameID
}

func (suite *E2ETestSuite) playCompleteGame(conn1, conn2 *websocket.Conn, gameID, player1, player2 string) string {
	// Play a simple winning pattern for player 1
	moves := []struct {
		conn   *websocket.Conn
		column int
	}{
		{conn1, 0}, {conn2, 1}, // Player1: col 0, Player2: col 1
		{conn1, 0}, {conn2, 1}, // Player1: col 0, Player2: col 1
		{conn1, 0}, {conn2, 1}, // Player1: col 0, Player2: col 1
		{conn1, 0},             // Player1: col 0 - WINS (4 in a row vertically)
	}
	
	for _, move := range moves {
		suite.makeMove(move.conn, gameID, move.column)
		
		// Read response
		response := suite.readWebSocketMessage(move.conn)
		
		// Check if game ended
		if response["type"] == "game_ended" {
			payload := response["payload"].(map[string]interface{})
			return payload["winner"].(string)
		}
	}
	
	return player1 // Default winner
}

func (suite *E2ETestSuite) playAgainstBot(conn *websocket.Conn, gameID, player string) {
	// Make a few moves against bot
	for i := 0; i < 3; i++ {
		suite.makeMove(conn, gameID, i)
		
		// Read player move response
		suite.readWebSocketMessage(conn)
		
		// Read bot move response (bot should respond automatically)
		botResponse := suite.readWebSocketMessage(conn)
		assert.Equal(suite.T(), "move_made", botResponse["type"])
		
		// Check if game ended
		if botResponse["type"] == "game_ended" {
			break
		}
	}
}

func (suite *E2ETestSuite) makeMove(conn *websocket.Conn, gameID string, column int) {
	moveMsg := map[string]interface{}{
		"type": "make_move",
		"payload": map[string]interface{}{
			"gameId": gameID,
			"column": column,
		},
	}
	suite.sendWebSocketMessage(conn, moveMsg)
}

func (suite *E2ETestSuite) sendWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	data, err := json.Marshal(message)
	require.NoError(suite.T(), err)
	
	err = conn.WriteMessage(websocket.TextMessage, data)
	require.NoError(suite.T(), err)
}

func (suite *E2ETestSuite) readWebSocketMessage(conn *websocket.Conn) map[string]interface{} {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	_, data, err := conn.ReadMessage()
	require.NoError(suite.T(), err)
	
	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	require.NoError(suite.T(), err)
	
	return response
}

func (suite *E2ETestSuite) verifyGameCompletion(gameID, winner string) {
	// Verify game session is completed
	session, err := suite.gameService.GetSession(suite.ctx, gameID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.GameStatusCompleted, session.Status)
	assert.NotNil(suite.T(), session.Winner)
	assert.Equal(suite.T(), winner, string(*session.Winner))
}

func (suite *E2ETestSuite) verifyBotGameCompletion(gameID string) {
	// Verify bot game was completed
	session, err := suite.gameService.GetSession(suite.ctx, gameID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), session.Status == models.GameStatusCompleted || session.Status == models.GameStatusInProgress)
}

func (suite *E2ETestSuite) verifyAnalyticsEvents(gameID, player1, player2 string) {
	// Wait for events to be processed
	time.Sleep(1 * time.Second)
	
	// Check that game events were created
	var events []models.GameEvent
	err := suite.db.Where("game_id = ?", gameID).Find(&events).Error
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(events), 0, "Game events should be recorded")
	
	// Verify event types
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.EventType] = true
	}
	
	assert.True(suite.T(), eventTypes["game_started"], "Should have game_started event")
	assert.True(suite.T(), eventTypes["move_made"], "Should have move_made events")
}

func (suite *E2ETestSuite) verifyLeaderboardUpdates(player1, player2, winner string) {
	// Check player statistics were updated
	stats1, err := suite.statsService.GetPlayerStats(suite.ctx, player1)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), stats1.GamesPlayed, 0)
	
	stats2, err := suite.statsService.GetPlayerStats(suite.ctx, player2)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), stats2.GamesPlayed, 0)
	
	// Check winner has increased win count
	if winner == player1 {
		assert.Greater(suite.T(), stats1.GamesWon, 0)
	} else if winner == player2 {
		assert.Greater(suite.T(), stats2.GamesWon, 0)
	}
	
	// Test leaderboard API
	url := fmt.Sprintf("%s/api/v1/leaderboard", suite.server.URL)
	resp, err := http.Get(url)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

// TestE2ETestSuite runs the end-to-end test suite
func TestE2ETestSuite(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	// Check if required environment variables are set
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration tests")
	}
	
	suite.Run(t, new(E2ETestSuite))
}