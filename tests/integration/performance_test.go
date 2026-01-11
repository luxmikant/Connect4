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
	"sync"
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

// PerformanceTestSuite provides performance and load testing
type PerformanceTestSuite struct {
	suite.Suite
	
	// Configuration
	config *config.Config
	
	// Database
	db         *gorm.DB
	repoManager *repositories.Manager
	
	// Services
	gameService        game.GameService
	matchmakingService matchmaking.Service
	statsService       stats.PlayerStatsService
	websocketService   *wsService.Service
	botService         bot.Service
	
	// HTTP Server
	router *gin.Engine
	server *httptest.Server
	
	// Test context
	ctx    context.Context
	cancel context.CancelFunc
	
	// Performance metrics
	metrics *PerformanceMetrics
}

// PerformanceMetrics tracks performance data during tests
type PerformanceMetrics struct {
	mu sync.RWMutex
	
	// Connection metrics
	ConnectionCount    int
	MaxConnections     int
	ConnectionErrors   int
	
	// Response time metrics
	ResponseTimes      []time.Duration
	MaxResponseTime    time.Duration
	MinResponseTime    time.Duration
	AvgResponseTime    time.Duration
	
	// Game metrics
	GamesCreated       int
	GamesCompleted     int
	GameErrors         int
	
	// Bot metrics
	BotMoves           int
	BotResponseTimes   []time.Duration
	BotTimeouts        int
	
	// Database metrics
	DatabaseQueries    int
	DatabaseErrors     int
	DatabaseLatency    []time.Duration
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		ResponseTimes:    make([]time.Duration, 0),
		BotResponseTimes: make([]time.Duration, 0),
		DatabaseLatency:  make([]time.Duration, 0),
		MinResponseTime:  time.Hour, // Initialize to high value
	}
}

// RecordResponseTime records a response time measurement
func (m *PerformanceMetrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ResponseTimes = append(m.ResponseTimes, duration)
	
	if duration > m.MaxResponseTime {
		m.MaxResponseTime = duration
	}
	
	if duration < m.MinResponseTime {
		m.MinResponseTime = duration
	}
	
	// Calculate running average
	total := time.Duration(0)
	for _, rt := range m.ResponseTimes {
		total += rt
	}
	m.AvgResponseTime = total / time.Duration(len(m.ResponseTimes))
}

// RecordBotResponseTime records a bot response time measurement
func (m *PerformanceMetrics) RecordBotResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.BotMoves++
	m.BotResponseTimes = append(m.BotResponseTimes, duration)
	
	if duration > time.Second {
		m.BotTimeouts++
	}
}

// IncrementConnections increments the connection count
func (m *PerformanceMetrics) IncrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ConnectionCount++
	if m.ConnectionCount > m.MaxConnections {
		m.MaxConnections = m.ConnectionCount
	}
}

// DecrementConnections decrements the connection count
func (m *PerformanceMetrics) DecrementConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ConnectionCount--
}

// GetSummary returns a performance summary
func (m *PerformanceMetrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	avgBotTime := time.Duration(0)
	if len(m.BotResponseTimes) > 0 {
		total := time.Duration(0)
		for _, bt := range m.BotResponseTimes {
			total += bt
		}
		avgBotTime = total / time.Duration(len(m.BotResponseTimes))
	}
	
	return map[string]interface{}{
		"max_connections":     m.MaxConnections,
		"connection_errors":   m.ConnectionErrors,
		"avg_response_time":   m.AvgResponseTime,
		"max_response_time":   m.MaxResponseTime,
		"min_response_time":   m.MinResponseTime,
		"games_created":       m.GamesCreated,
		"games_completed":     m.GamesCompleted,
		"game_errors":         m.GameErrors,
		"bot_moves":           m.BotMoves,
		"avg_bot_time":        avgBotTime,
		"bot_timeouts":        m.BotTimeouts,
		"database_queries":    m.DatabaseQueries,
		"database_errors":     m.DatabaseErrors,
	}
}

// SetupSuite initializes the performance test environment
func (suite *PerformanceTestSuite) SetupSuite() {
	// Initialize metrics
	suite.metrics = NewPerformanceMetrics()
	
	// Set up test context
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	
	// Load configuration
	cfg, err := config.Load()
	require.NoError(suite.T(), err, "Failed to load configuration")
	suite.config = cfg
	
	// Override config for performance testing
	suite.setupPerformanceConfig()
	
	// Initialize database
	suite.setupDatabase()
	
	// Initialize services
	suite.setupServices()
	
	// Initialize HTTP server
	suite.setupHTTPServer()
	
	// Start services
	suite.startServices()
}

// TearDownSuite cleans up the performance test environment
func (suite *PerformanceTestSuite) TearDownSuite() {
	// Print performance summary
	summary := suite.metrics.GetSummary()
	suite.T().Logf("Performance Test Summary: %+v", summary)
	
	// Stop services
	if suite.websocketService != nil {
		suite.websocketService.Stop()
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

// setupPerformanceConfig configures the environment for performance testing
func (suite *PerformanceTestSuite) setupPerformanceConfig() {
	// Use performance-optimized database settings
	suite.config.Database.MaxOpenConns = 50
	suite.config.Database.MaxIdleConns = 10
	
	// Use test database if available
	if testDB := os.Getenv("TEST_DATABASE_URL"); testDB != "" {
		suite.config.Database.URL = testDB
	}
	
	// Set performance environment
	suite.config.Environment = "performance"
}

// setupDatabase initializes the database for performance testing
func (suite *PerformanceTestSuite) setupDatabase() {
	// Connect to database
	db, err := gorm.Open(postgres.Open(suite.config.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce noise in tests
	})
	require.NoError(suite.T(), err, "Failed to connect to database")
	
	suite.db = db
	
	// Configure connection pool for performance
	sqlDB, err := db.DB()
	require.NoError(suite.T(), err)
	
	sqlDB.SetMaxOpenConns(suite.config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(suite.config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(suite.config.Database.ConnMaxLifetime) * time.Second)
	
	// Run migrations
	migrator := database.NewMigrator(db)
	err = migrator.Up()
	require.NoError(suite.T(), err, "Failed to run database migrations")
	
	// Create repository manager
	suite.repoManager = repositories.NewManager(db)
	
	// Clean existing test data
	suite.cleanupTestData()
}

// setupServices initializes services for performance testing
func (suite *PerformanceTestSuite) setupServices() {
	// Game service
	suite.gameService = game.NewService(suite.repoManager, slog.Default())
	
	// Bot service
	suite.botService = bot.NewService(slog.Default())
	
	// Stats service
	suite.statsService = stats.NewService(suite.repoManager, slog.Default())
	
	// Matchmaking service with optimized settings
	matchmakingConfig := &matchmaking.ServiceConfig{
		MatchTimeout:  1 * time.Second, // Fast matching for performance tests
		MatchInterval: 50 * time.Millisecond,
		Logger:        slog.Default(),
	}
	suite.matchmakingService = matchmaking.NewMatchmakingService(suite.gameService, matchmakingConfig)
	
	// WebSocket service
	suite.websocketService = wsService.NewService(suite.gameService, suite.matchmakingService)
}

// setupHTTPServer initializes the HTTP server for performance testing
func (suite *PerformanceTestSuite) setupHTTPServer() {
	gin.SetMode(gin.ReleaseMode) // Use release mode for better performance
	
	// Create router
	suite.router = gin.New()
	
	// Add minimal middleware for performance
	suite.router.Use(middleware.Recovery())
	
	// Create handlers
	gameHandler := handlers.NewGameHandler(suite.gameService, suite.botService, suite.statsService)
	leaderboardHandler := handlers.NewLeaderboardHandler(suite.statsService)
	
	// Setup routes
	routes.SetupRoutes(suite.router, gameHandler, leaderboardHandler, suite.websocketService.GetWebSocketHandler())
	
	// Create test server
	suite.server = httptest.NewServer(suite.router)
}

// startServices starts all services for performance testing
func (suite *PerformanceTestSuite) startServices() {
	// Start WebSocket service
	err := suite.websocketService.Start(suite.ctx)
	require.NoError(suite.T(), err, "Failed to start WebSocket service")
	
	// Allow services to initialize
	time.Sleep(100 * time.Millisecond)
}

// TestConcurrentGameSessions tests multiple concurrent game sessions
func (suite *PerformanceTestSuite) TestConcurrentGameSessions() {
	numGames := 10
	concurrency := 20 // 20 concurrent connections (10 games)
	
	var wg sync.WaitGroup
	gameResults := make(chan string, numGames)
	errors := make(chan error, concurrency)
	
	// Create concurrent games
	for i := 0; i < numGames; i++ {
		wg.Add(2) // Two players per game
		
		go func(gameIndex int) {
			defer wg.Done()
			
			player1 := fmt.Sprintf("perf_p1_%d", gameIndex)
			player2 := fmt.Sprintf("perf_p2_%d", gameIndex)
			
			// Create players
			suite.createPlayerConcurrent(player1, errors)
			suite.createPlayerConcurrent(player2, errors)
			
			// Setup WebSocket connections
			conn1, err := suite.setupWebSocketConnectionConcurrent(player1)
			if err != nil {
				errors <- err
				return
			}
			defer conn1.Close()
			
			suite.metrics.IncrementConnections()
			defer suite.metrics.DecrementConnections()
			
			// Measure response time
			start := time.Now()
			
			// Wait for game start
			response := suite.readWebSocketMessageWithTimeout(conn1, 5*time.Second)
			if response == nil {
				errors <- fmt.Errorf("timeout waiting for game start")
				return
			}
			
			responseTime := time.Since(start)
			suite.metrics.RecordResponseTime(responseTime)
			
			if response["type"] == "game_started" {
				payload := response["payload"].(map[string]interface{})
				gameID := payload["gameId"].(string)
				gameResults <- gameID
				
				suite.metrics.mu.Lock()
				suite.metrics.GamesCreated++
				suite.metrics.mu.Unlock()
			}
		}(i)
		
		go func(gameIndex int) {
			defer wg.Done()
			
			player2 := fmt.Sprintf("perf_p2_%d", gameIndex)
			
			// Setup second connection
			conn2, err := suite.setupWebSocketConnectionConcurrent(player2)
			if err != nil {
				errors <- err
				return
			}
			defer conn2.Close()
			
			suite.metrics.IncrementConnections()
			defer suite.metrics.DecrementConnections()
			
			// Wait for game start
			response := suite.readWebSocketMessageWithTimeout(conn2, 5*time.Second)
			if response != nil && response["type"] == "game_started" {
				// Game started successfully
			}
		}(i)
	}
	
	// Wait for all games to start
	wg.Wait()
	close(gameResults)
	close(errors)
	
	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			suite.T().Logf("Concurrent game error: %v", err)
			errorCount++
		}
	}
	
	// Collect game IDs
	gameIDs := make([]string, 0, numGames)
	for gameID := range gameResults {
		gameIDs = append(gameIDs, gameID)
	}
	
	// Verify performance requirements
	assert.LessOrEqual(suite.T(), errorCount, numGames/10, "Error rate should be less than 10%")
	assert.GreaterOrEqual(suite.T(), len(gameIDs), numGames*8/10, "At least 80% of games should start successfully")
	assert.LessOrEqual(suite.T(), suite.metrics.AvgResponseTime, 2*time.Second, "Average response time should be under 2 seconds")
	
	suite.T().Logf("Created %d games with %d errors, avg response time: %v", 
		len(gameIDs), errorCount, suite.metrics.AvgResponseTime)
}

// TestWebSocketPerformance tests WebSocket performance under load
func (suite *PerformanceTestSuite) TestWebSocketPerformance() {
	numConnections := 50
	messagesPerConnection := 10
	
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, numConnections)
	errors := make(chan error, numConnections*messagesPerConnection)
	
	// Create connections
	for i := 0; i < numConnections; i++ {
		player := fmt.Sprintf("ws_perf_%d", i)
		suite.createPlayerConcurrent(player, errors)
		
		conn, err := suite.setupWebSocketConnectionConcurrent(player)
		if err != nil {
			suite.T().Errorf("Failed to create connection %d: %v", i, err)
			continue
		}
		connections[i] = conn
		suite.metrics.IncrementConnections()
	}
	
	// Send messages concurrently
	start := time.Now()
	
	for i, conn := range connections {
		if conn == nil {
			continue
		}
		
		wg.Add(1)
		go func(connIndex int, connection *websocket.Conn) {
			defer wg.Done()
			defer connection.Close()
			defer suite.metrics.DecrementConnections()
			
			for j := 0; j < messagesPerConnection; j++ {
				msgStart := time.Now()
				
				// Send ping message
				pingMsg := map[string]interface{}{
					"type": "ping",
					"payload": map[string]interface{}{
						"timestamp": time.Now().Unix(),
					},
				}
				
				err := suite.sendWebSocketMessageConcurrent(connection, pingMsg)
				if err != nil {
					errors <- err
					continue
				}
				
				// Read response
				response := suite.readWebSocketMessageWithTimeout(connection, 1*time.Second)
				if response == nil {
					errors <- fmt.Errorf("timeout reading message response")
					continue
				}
				
				responseTime := time.Since(msgStart)
				suite.metrics.RecordResponseTime(responseTime)
			}
		}(i, conn)
	}
	
	wg.Wait()
	totalTime := time.Since(start)
	close(errors)
	
	// Count errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}
	
	totalMessages := numConnections * messagesPerConnection
	successRate := float64(totalMessages-errorCount) / float64(totalMessages) * 100
	
	// Performance assertions
	assert.GreaterOrEqual(suite.T(), successRate, 95.0, "Success rate should be at least 95%")
	assert.LessOrEqual(suite.T(), suite.metrics.AvgResponseTime, 100*time.Millisecond, "Average response time should be under 100ms")
	assert.LessOrEqual(suite.T(), suite.metrics.MaxConnections, numConnections+5, "Connection count should be within expected range")
	
	suite.T().Logf("WebSocket Performance: %d connections, %d messages, %.2f%% success rate, %v total time, %v avg response",
		numConnections, totalMessages, successRate, totalTime, suite.metrics.AvgResponseTime)
}

// TestBotResponseTimes tests bot AI performance under load
func (suite *PerformanceTestSuite) TestBotResponseTimes() {
	numBotGames := 20
	
	var wg sync.WaitGroup
	botTimes := make(chan time.Duration, numBotGames*10) // Estimate 10 moves per game
	
	for i := 0; i < numBotGames; i++ {
		wg.Add(1)
		
		go func(gameIndex int) {
			defer wg.Done()
			
			player := fmt.Sprintf("bot_perf_%d", gameIndex)
			suite.createPlayerConcurrent(player, nil)
			
			conn, err := suite.setupWebSocketConnectionConcurrent(player)
			if err != nil {
				suite.T().Errorf("Failed to create bot game connection: %v", err)
				return
			}
			defer conn.Close()
			
			// Wait for bot game to start (should timeout and create bot game)
			time.Sleep(2 * time.Second)
			
			response := suite.readWebSocketMessageWithTimeout(conn, 3*time.Second)
			if response == nil || response["type"] != "game_started" {
				return
			}
			
			payload := response["payload"].(map[string]interface{})
			gameID := payload["gameId"].(string)
			
			// Play against bot and measure response times
			for move := 0; move < 5; move++ {
				// Make player move
				suite.makeMoveConcurrent(conn, gameID, move%7)
				
				// Read player move response
				suite.readWebSocketMessageWithTimeout(conn, 1*time.Second)
				
				// Measure bot response time
				botStart := time.Now()
				botResponse := suite.readWebSocketMessageWithTimeout(conn, 2*time.Second)
				botTime := time.Since(botStart)
				
				if botResponse != nil && botResponse["type"] == "move_made" {
					botTimes <- botTime
					suite.metrics.RecordBotResponseTime(botTime)
				}
				
				// Check if game ended
				if botResponse != nil && botResponse["type"] == "game_ended" {
					break
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(botTimes)
	
	// Analyze bot response times
	var totalBotTime time.Duration
	var maxBotTime time.Duration
	var minBotTime time.Duration = time.Hour
	botMoveCount := 0
	timeoutCount := 0
	
	for botTime := range botTimes {
		totalBotTime += botTime
		botMoveCount++
		
		if botTime > maxBotTime {
			maxBotTime = botTime
		}
		if botTime < minBotTime {
			minBotTime = botTime
		}
		if botTime > time.Second {
			timeoutCount++
		}
	}
	
	avgBotTime := time.Duration(0)
	if botMoveCount > 0 {
		avgBotTime = totalBotTime / time.Duration(botMoveCount)
	}
	
	// Performance assertions for bot
	assert.Greater(suite.T(), botMoveCount, 0, "Should have recorded bot moves")
	assert.LessOrEqual(suite.T(), avgBotTime, 800*time.Millisecond, "Average bot response time should be under 800ms")
	assert.LessOrEqual(suite.T(), maxBotTime, 1200*time.Millisecond, "Max bot response time should be under 1.2s")
	assert.LessOrEqual(suite.T(), timeoutCount, botMoveCount/10, "Bot timeout rate should be less than 10%")
	
	suite.T().Logf("Bot Performance: %d moves, avg: %v, max: %v, min: %v, timeouts: %d",
		botMoveCount, avgBotTime, maxBotTime, minBotTime, timeoutCount)
}

// TestDatabasePerformance tests database performance under load
func (suite *PerformanceTestSuite) TestDatabasePerformance() {
	numQueries := 100
	concurrency := 10
	
	var wg sync.WaitGroup
	queryTimes := make(chan time.Duration, numQueries)
	errors := make(chan error, numQueries)
	
	// Create test data
	for i := 0; i < 20; i++ {
		player := &models.Player{
			Username: fmt.Sprintf("db_perf_%d", i),
		}
		suite.repoManager.Player().Create(suite.ctx, player)
	}
	
	// Run concurrent database queries
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		
		go func(workerID int) {
			defer wg.Done()
			
			queriesPerWorker := numQueries / concurrency
			
			for j := 0; j < queriesPerWorker; j++ {
				start := time.Now()
				
				// Test different types of queries
				switch j % 4 {
				case 0:
					// Player lookup
					_, err := suite.repoManager.Player().GetByUsername(suite.ctx, fmt.Sprintf("db_perf_%d", j%20))
					if err != nil {
						errors <- err
					}
				case 1:
					// Stats query
					_, err := suite.repoManager.PlayerStats().GetByUsername(suite.ctx, fmt.Sprintf("db_perf_%d", j%20))
					if err != nil && err != gorm.ErrRecordNotFound {
						errors <- err
					}
				case 2:
					// Game session query
					_, err := suite.repoManager.GameSession().GetActiveSessions(suite.ctx)
					if err != nil {
						errors <- err
					}
				case 3:
					// Complex join query
					_, err := suite.repoManager.GameSession().GetSessionsByPlayer(suite.ctx, fmt.Sprintf("db_perf_%d", j%20))
					if err != nil {
						errors <- err
					}
				}
				
				queryTime := time.Since(start)
				queryTimes <- queryTime
				
				suite.metrics.mu.Lock()
				suite.metrics.DatabaseQueries++
				suite.metrics.DatabaseLatency = append(suite.metrics.DatabaseLatency, queryTime)
				suite.metrics.mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	close(queryTimes)
	close(errors)
	
	// Analyze database performance
	var totalQueryTime time.Duration
	var maxQueryTime time.Duration
	var minQueryTime time.Duration = time.Hour
	queryCount := 0
	
	for queryTime := range queryTimes {
		totalQueryTime += queryTime
		queryCount++
		
		if queryTime > maxQueryTime {
			maxQueryTime = queryTime
		}
		if queryTime < minQueryTime {
			minQueryTime = queryTime
		}
	}
	
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}
	
	avgQueryTime := time.Duration(0)
	if queryCount > 0 {
		avgQueryTime = totalQueryTime / time.Duration(queryCount)
	}
	
	// Performance assertions for database
	assert.LessOrEqual(suite.T(), errorCount, numQueries/20, "Database error rate should be less than 5%")
	assert.LessOrEqual(suite.T(), avgQueryTime, 50*time.Millisecond, "Average query time should be under 50ms")
	assert.LessOrEqual(suite.T(), maxQueryTime, 200*time.Millisecond, "Max query time should be under 200ms")
	
	suite.T().Logf("Database Performance: %d queries, avg: %v, max: %v, min: %v, errors: %d",
		queryCount, avgQueryTime, maxQueryTime, minQueryTime, errorCount)
}

// TestSupabaseConnectionLimits tests Supabase connection pool limits
func (suite *PerformanceTestSuite) TestSupabaseConnectionLimits() {
	// Test connection pool behavior
	maxConnections := suite.config.Database.MaxOpenConns
	
	// Create more connections than the pool limit
	var wg sync.WaitGroup
	connectionResults := make(chan bool, maxConnections*2)
	
	for i := 0; i < maxConnections*2; i++ {
		wg.Add(1)
		
		go func(connIndex int) {
			defer wg.Done()
			
			// Try to perform a database operation
			start := time.Now()
			player := &models.Player{
				Username: fmt.Sprintf("conn_test_%d_%d", connIndex, time.Now().UnixNano()),
			}
			
			err := suite.repoManager.Player().Create(suite.ctx, player)
			duration := time.Since(start)
			
			if err != nil {
				connectionResults <- false
				suite.metrics.mu.Lock()
				suite.metrics.DatabaseErrors++
				suite.metrics.mu.Unlock()
			} else {
				connectionResults <- true
				suite.metrics.mu.Lock()
				suite.metrics.DatabaseLatency = append(suite.metrics.DatabaseLatency, duration)
				suite.metrics.mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	close(connectionResults)
	
	// Analyze connection results
	successCount := 0
	failureCount := 0
	
	for success := range connectionResults {
		if success {
			successCount++
		} else {
			failureCount++
		}
	}
	
	successRate := float64(successCount) / float64(successCount+failureCount) * 100
	
	// Verify connection pool handles load gracefully
	assert.GreaterOrEqual(suite.T(), successRate, 90.0, "Connection success rate should be at least 90%")
	assert.LessOrEqual(suite.T(), failureCount, maxConnections/5, "Connection failures should be limited")
	
	suite.T().Logf("Connection Pool Test: %d success, %d failures, %.2f%% success rate",
		successCount, failureCount, successRate)
}

// Helper methods for concurrent testing

func (suite *PerformanceTestSuite) createPlayerConcurrent(username string, errors chan<- error) {
	url := fmt.Sprintf("%s/api/v1/players", suite.server.URL)
	payload := fmt.Sprintf(`{"username": "%s"}`, username)
	
	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil && errors != nil {
		errors <- err
		return
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func (suite *PerformanceTestSuite) setupWebSocketConnectionConcurrent(username string) (*websocket.Conn, error) {
	wsURL := strings.Replace(suite.server.URL, "http://", "ws://", 1) + "/ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}
	
	joinMsg := map[string]interface{}{
		"type": "join_game",
		"payload": map[string]interface{}{
			"username": username,
		},
	}
	
	err = suite.sendWebSocketMessageConcurrent(conn, joinMsg)
	if err != nil {
		conn.Close()
		return nil, err
	}
	
	return conn, nil
}

func (suite *PerformanceTestSuite) sendWebSocketMessageConcurrent(conn *websocket.Conn, message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (suite *PerformanceTestSuite) readWebSocketMessageWithTimeout(conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	conn.SetReadDeadline(time.Now().Add(timeout))
	
	_, data, err := conn.ReadMessage()
	if err != nil {
		return nil
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil
	}
	
	return response
}

func (suite *PerformanceTestSuite) makeMoveConcurrent(conn *websocket.Conn, gameID string, column int) {
	moveMsg := map[string]interface{}{
		"type": "make_move",
		"payload": map[string]interface{}{
			"gameId": gameID,
			"column": column,
		},
	}
	suite.sendWebSocketMessageConcurrent(conn, moveMsg)
}

func (suite *PerformanceTestSuite) cleanupDatabase() {
	if suite.db != nil {
		suite.cleanupTestData()
	}
}

func (suite *PerformanceTestSuite) cleanupTestData() {
	// Delete test data
	suite.db.Exec("DELETE FROM game_events WHERE game_id LIKE 'perf_%' OR game_id LIKE 'ws_%' OR game_id LIKE 'bot_%' OR game_id LIKE 'conn_%'")
	suite.db.Exec("DELETE FROM moves WHERE game_id LIKE 'perf_%' OR game_id LIKE 'ws_%' OR game_id LIKE 'bot_%' OR game_id LIKE 'conn_%'")
	suite.db.Exec("DELETE FROM game_sessions WHERE id LIKE 'perf_%' OR id LIKE 'ws_%' OR id LIKE 'bot_%' OR id LIKE 'conn_%'")
	suite.db.Exec("DELETE FROM player_stats WHERE username LIKE 'perf_%' OR username LIKE 'ws_%' OR username LIKE 'bot_%' OR username LIKE 'conn_%' OR username LIKE 'db_%'")
	suite.db.Exec("DELETE FROM players WHERE username LIKE 'perf_%' OR username LIKE 'ws_%' OR username LIKE 'bot_%' OR username LIKE 'conn_%' OR username LIKE 'db_%'")
}

// TestPerformanceTestSuite runs the performance test suite
func TestPerformanceTestSuite(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}
	
	// Check if required environment variables are set
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping performance tests")
	}
	
	suite.Run(t, new(PerformanceTestSuite))
}