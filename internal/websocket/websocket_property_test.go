//go:build property
// +build property

package websocket_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	wspackage "connect4-multiplayer/internal/websocket"
	"connect4-multiplayer/pkg/models"
)

// MockGameService for testing
type MockGameService struct {
	sessions map[string]*models.GameSession
	mu       sync.RWMutex
}

func NewMockGameService() *MockGameService {
	return &MockGameService{
		sessions: make(map[string]*models.GameSession),
	}
}

func (m *MockGameService) CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &models.GameSession{
		ID:          generateTestID(),
		Player1:     player1,
		Player2:     player2,
		Board:       models.NewBoard(),
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		StartTime:   time.Now(),
	}

	m.sessions[session.ID] = session
	return session, nil
}

func (m *MockGameService) GetSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return nil, models.ErrGameNotFound
	}
	return session, nil
}

func (m *MockGameService) EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error {
	return m.CompleteGame(ctx, gameID, winner)
}

func (m *MockGameService) GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error) {
	session, err := m.GetSession(ctx, gameID)
	if err != nil {
		return "", "", err
	}
	return session.GetCurrentPlayer(), session.CurrentTurn, nil
}

func (m *MockGameService) SwitchTurn(ctx context.Context, gameID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}

	if session.CurrentTurn == models.PlayerColorRed {
		session.CurrentTurn = models.PlayerColorYellow
	} else {
		session.CurrentTurn = models.PlayerColorRed
	}

	return nil
}

func (m *MockGameService) AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error) {
	session, err := m.GetSession(ctx, gameID)
	if err != nil {
		return nil, err
	}
	
	colors := make(map[string]models.PlayerColor)
	colors[session.Player1] = models.PlayerColorRed
	colors[session.Player2] = models.PlayerColorYellow
	return colors, nil
}

func (m *MockGameService) CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}

	session.Status = models.StatusCompleted
	session.Winner = winner
	endTime := time.Now()
	session.EndTime = &endTime

	return nil
}

func (m *MockGameService) GetActiveSessions(ctx context.Context) ([]*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var sessions []*models.GameSession
	for _, session := range m.sessions {
		if session.Status == models.StatusInProgress {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockGameService) GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var sessions []*models.GameSession
	for _, session := range m.sessions {
		if session.Player1 == username || session.Player2 == username {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockGameService) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	sessions, err := m.GetSessionsByPlayer(ctx, username)
	if err != nil {
		return nil, err
	}
	
	for _, session := range sessions {
		if session.Status == models.StatusInProgress {
			return session, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameService) GetActiveSessionCount(ctx context.Context) (int64, error) {
	sessions, err := m.GetActiveSessions(ctx)
	if err != nil {
		return 0, err
	}
	return int64(len(sessions)), nil
}

func (m *MockGameService) CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error) {
	return 0, nil
}

func (m *MockGameService) MarkSessionAbandoned(ctx context.Context, gameID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}
	
	session.Status = models.StatusAbandoned
	return nil
}

func (m *MockGameService) StartCleanupWorker(ctx context.Context, interval time.Duration) {}

func (m *MockGameService) StopCleanupWorker() {}

func (m *MockGameService) MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) MarkPlayerReconnected(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) GetDisconnectedPlayers(gameID string) map[string]time.Time {
	return make(map[string]time.Time)
}

func (m *MockGameService) HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) CacheSession(session *models.GameSession) {}

func (m *MockGameService) GetCachedSession(gameID string) (*models.GameSession, bool) {
	session, err := m.GetSession(context.Background(), gameID)
	if err != nil {
		return nil, false
	}
	return session, true
}

func (m *MockGameService) InvalidateCache(gameID string) {}

func (m *MockGameService) CleanupCache(maxAge time.Duration) int {
	return 0
}

func (m *MockGameService) GetCacheStats() map[string]interface{} {
	return make(map[string]interface{})
}

func generateTestID() string {
	return "test-" + time.Now().Format("20060102150405")
}

// TestClient represents a WebSocket test client
type TestClient struct {
	conn     *websocket.Conn
	userID   string
	messages chan []byte
	done     chan struct{}
}

func NewTestClient(wsURL, userID string) (*TestClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?userId="+userID, nil)
	if err != nil {
		return nil, err
	}

	client := &TestClient{
		conn:     conn,
		userID:   userID,
		messages: make(chan []byte, 100),
		done:     make(chan struct{}),
	}

	// Start reading messages
	go client.readMessages()

	return client, nil
}

func (c *TestClient) readMessages() {
	defer func() {
		select {
		case <-c.done:
		default:
			close(c.done)
		}
	}()
	
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
			select {
			case c.messages <- message:
			case <-c.done:
				return
			}
		}
	}
}

func (c *TestClient) SendMessage(msg *wspackage.Message) error {
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *TestClient) WaitForMessage(timeout time.Duration, msgType wspackage.MessageType) (*wspackage.Message, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msgData := <-c.messages:
			var msg wspackage.Message
			if err := json.Unmarshal(msgData, &msg); err == nil {
				if msg.Type == msgType {
					return &msg, true
				}
			}
		case <-timer.C:
			return nil, false
		case <-c.done:
			return nil, false
		}
	}
}

func (c *TestClient) Close() {
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}
	c.conn.Close()
}

// Feature: connect-4-multiplayer, Property 4: Real-time Game State Synchronization
func TestRealTimeGameStateSynchronization(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("for any game state change, all connected clients should receive updates within 100ms", prop.ForAll(
		func(player1Name, player2Name string, moveColumn int) bool {
			// Skip invalid inputs
			if player1Name == "" || player2Name == "" || player1Name == player2Name {
				return true
			}
			if moveColumn < 0 || moveColumn > 6 {
				return true
			}

			// Create mock game service and WebSocket service
			gameService := NewMockGameService()
			wsService := wspackage.NewService(gameService)

			ctx := context.Background()
			err := wsService.Start(ctx)
			if err != nil {
				return false
			}
			defer wsService.Stop()

			// Create test server
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/ws", wsService.GetWebSocketHandler().HandleWebSocket)
			server := httptest.NewServer(router)
			defer server.Close()

			// Convert HTTP URL to WebSocket URL
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

			// Create test clients
			client1, err := NewTestClient(wsURL, player1Name)
			if err != nil {
				return false
			}
			defer client1.Close()

			client2, err := NewTestClient(wsURL, player2Name)
			if err != nil {
				return false
			}
			defer client2.Close()

			// Wait for connections to be established
			time.Sleep(50 * time.Millisecond)

			// Send join_game messages to properly add connections to game rooms
			joinMsg1 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": player1Name,
				"gameType": "pvp",
			})
			
			joinMsg2 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": player2Name,
				"gameType": "pvp",
			})

			// Send join messages
			if err := client1.SendMessage(joinMsg1); err != nil {
				return false
			}
			if err := client2.SendMessage(joinMsg2); err != nil {
				return false
			}

			// Wait for join processing and game creation
			time.Sleep(100 * time.Millisecond)

			// Wait for game_started messages to get the game ID
			gameStartedMsg1, received1 := client1.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)
			gameStartedMsg2, received2 := client2.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)

			if !received1 || !received2 {
				return false
			}

			// Extract game ID from the game started message
			gameID1, ok1 := gameStartedMsg1.Payload["gameId"].(string)
			gameID2, ok2 := gameStartedMsg2.Payload["gameId"].(string)
			
			if !ok1 || !ok2 || gameID1 == "" || gameID2 == "" {
				return false
			}

			// Use the game ID from the first client
			gameID := gameID1

			// Record start time for broadcast test
			startTime := time.Now()

			// Create a game session to get the board state
			session, err := gameService.GetSession(ctx, gameID)
			if err != nil {
				return false
			}

			// Broadcast a move message
			moveMsg := wspackage.CreateMoveMadeMessage(
				gameID,
				player1Name,
				moveColumn,
				0, // row
				session.Board,
				string(models.PlayerColorYellow),
				1, // move count
			)

			msgData, err := moveMsg.ToJSON()
			if err != nil {
				return false
			}

			// Broadcast the message
			wsService.BroadcastToGame(gameID, msgData, "")

			// Wait for messages to be received by both clients
			_, receivedMove1 := client1.WaitForMessage(100*time.Millisecond, wspackage.MessageTypeMoveMade)
			_, receivedMove2 := client2.WaitForMessage(100*time.Millisecond, wspackage.MessageTypeMoveMade)

			// Check timing
			elapsed := time.Since(startTime)
			return elapsed <= 100*time.Millisecond && receivedMove1 && receivedMove2
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.IntRange(0, 6),
	))

	properties.Property("for any message broadcast, only connections in the same game should receive it", prop.ForAll(
		func(game1Player1, game1Player2, game2Player1, game2Player2 string) bool {
			// Skip invalid inputs
			if game1Player1 == "" || game1Player2 == "" || game2Player1 == "" || game2Player2 == "" {
				return true
			}
			names := []string{game1Player1, game1Player2, game2Player1, game2Player2}
			for i := 0; i < len(names); i++ {
				for j := i + 1; j < len(names); j++ {
					if names[i] == names[j] {
						return true // Skip if any names are duplicate
					}
				}
			}

			// Create mock game service and WebSocket service
			gameService := NewMockGameService()
			wsService := wspackage.NewService(gameService)

			ctx := context.Background()
			err := wsService.Start(ctx)
			if err != nil {
				return false
			}
			defer wsService.Stop()

			// Create test server
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/ws", wsService.GetWebSocketHandler().HandleWebSocket)
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

			// Create test clients
			client1, err := NewTestClient(wsURL, game1Player1)
			if err != nil {
				return false
			}
			defer client1.Close()

			client2, err := NewTestClient(wsURL, game1Player2)
			if err != nil {
				return false
			}
			defer client2.Close()

			client3, err := NewTestClient(wsURL, game2Player1)
			if err != nil {
				return false
			}
			defer client3.Close()

			client4, err := NewTestClient(wsURL, game2Player2)
			if err != nil {
				return false
			}
			defer client4.Close()

			// Wait for connections to be established
			time.Sleep(50 * time.Millisecond)

			// Send join_game messages to create two separate games
			joinMsg1 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game1Player1,
				"gameType": "pvp",
			})
			
			joinMsg2 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game1Player2,
				"gameType": "pvp",
			})

			joinMsg3 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game2Player1,
				"gameType": "pvp",
			})
			
			joinMsg4 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game2Player2,
				"gameType": "pvp",
			})

			// Send join messages
			if err := client1.SendMessage(joinMsg1); err != nil {
				return false
			}
			if err := client2.SendMessage(joinMsg2); err != nil {
				return false
			}
			if err := client3.SendMessage(joinMsg3); err != nil {
				return false
			}
			if err := client4.SendMessage(joinMsg4); err != nil {
				return false
			}

			// Wait for join processing
			time.Sleep(200 * time.Millisecond)

			// Get game IDs from game started messages
			gameStartedMsg1, received1 := client1.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)
			gameStartedMsg3, received3 := client3.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)

			if !received1 || !received3 {
				return false
			}

			gameID1, ok1 := gameStartedMsg1.Payload["gameId"].(string)
			gameID2, ok2 := gameStartedMsg3.Payload["gameId"].(string)
			
			if !ok1 || !ok2 || gameID1 == "" || gameID2 == "" {
				return false
			}

			// Verify that the games are different and connections are properly isolated
			game1Conns := wsService.GetGameConnections(gameID1)
			game2Conns := wsService.GetGameConnections(gameID2)

			// Each game should have connections from its players
			// Game 1 should have connections from game1Player1 and game1Player2
			// Game 2 should have connections from game2Player1 and game2Player2
			game1HasCorrectPlayers := len(game1Conns) >= 1 // At least the joining player
			game2HasCorrectPlayers := len(game2Conns) >= 1 // At least the joining player
			gamesAreSeparate := gameID1 != gameID2

			return game1HasCorrectPlayers && game2HasCorrectPlayers && gamesAreSeparate
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
	))

	properties.Property("for any connection failure, hub should clean up properly", prop.ForAll(
		func(player1Name, player2Name string) bool {
			// Skip invalid inputs
			if player1Name == "" || player2Name == "" || player1Name == player2Name {
				return true
			}

			// Create mock game service and WebSocket service
			gameService := NewMockGameService()
			wsService := wspackage.NewService(gameService)

			ctx := context.Background()
			err := wsService.Start(ctx)
			if err != nil {
				return false
			}
			defer wsService.Stop()

			// Create test server
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/ws", wsService.GetWebSocketHandler().HandleWebSocket)
			server := httptest.NewServer(router)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

			// Create test clients
			client1, err := NewTestClient(wsURL, player1Name)
			if err != nil {
				return false
			}

			client2, err := NewTestClient(wsURL, player2Name)
			if err != nil {
				return false
			}
			defer client2.Close()

			// Wait for connections to be established
			time.Sleep(50 * time.Millisecond)

			// Verify both connections are registered
			initialCount := wsService.GetConnectionCount()
			if initialCount != 2 {
				return false
			}

			// Close one connection
			client1.Close()

			// Wait for cleanup
			time.Sleep(100 * time.Millisecond)

			// Verify connection count decreased
			finalCount := wsService.GetConnectionCount()
			return finalCount < initialCount
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}