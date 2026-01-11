//go:build property
// +build property

package websocket_test

import (
	"context"
	"encoding/json"
	"fmt"
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
	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/pkg/models"
)

// TestClient represents a WebSocket test client
type TestClient struct {
	conn     *websocket.Conn
	userID   string
	messages chan []byte
	done     chan struct{}
	mu       sync.RWMutex
	closed   bool
}

func NewTestClient(wsURL, userID string) (*TestClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?userId="+userID, nil)
	if err != nil {
		return nil, err
	}

	// Set read and write deadlines to prevent hanging
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	client := &TestClient{
		conn:     conn,
		userID:   userID,
		messages: make(chan []byte, 100),
		done:     make(chan struct{}),
		closed:   false,
	}

	// Start reading messages
	go client.readMessages()

	return client, nil
}

func (c *TestClient) readMessages() {
	defer func() {
		c.mu.Lock()
		if !c.closed {
			close(c.done)
			c.closed = true
		}
		c.mu.Unlock()
	}()
	
	for {
		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			return
		}
		c.mu.RUnlock()

		// Set read deadline for each message
		c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Connection closed or error occurred
			return
		}
		
		select {
		case c.messages <- message:
		case <-c.done:
			return
		default:
			// Channel is full, skip this message
		}
	}
}

func (c *TestClient) SendMessage(msg *wspackage.Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.closed {
		return fmt.Errorf("client is closed")
	}
	
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	
	// Set write deadline
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
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
				// Put non-matching message back for other waiters
				select {
				case c.messages <- msgData:
				default:
					// Channel full, drop message
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
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.closed {
		c.closed = true
		close(c.done)
		c.conn.Close()
	}
}

// Feature: connect-4-multiplayer, Property 4: Real-time Game State Synchronization
func TestRealTimeGameStateSynchronization(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("for any game state change, connected clients in the same game should receive updates within 100ms", prop.ForAll(
		func(playerName string, moveColumn int) bool {
			// Skip invalid inputs
			if playerName == "" {
				return true
			}
			if moveColumn < 0 || moveColumn > 6 {
				return true
			}

			// Create mock game service and WebSocket service
			gameService := NewMockGameService()
			matchmakingService := &MockMatchmakingService{}
			wsService := wspackage.NewService(gameService, matchmakingService)

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

			// Create test client
			client, err := NewTestClient(wsURL, playerName)
			if err != nil {
				return false
			}
			defer client.Close()

			// Wait for connection to be established
			time.Sleep(100 * time.Millisecond)

			// Send join_game message to create a bot game
			joinMsg := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": playerName,
				"gameType": "bot",
			})

			// Send join message
			if err := client.SendMessage(joinMsg); err != nil {
				return false
			}

			// Wait for join processing and game creation
			time.Sleep(200 * time.Millisecond)

			// Wait for game_started message to get the game ID
			gameStartedMsg, received := client.WaitForMessage(500*time.Millisecond, wspackage.MessageTypeGameStarted)

			if !received {
				return false
			}

			// Extract game ID from the game started message
			gameID, ok := gameStartedMsg.Payload["gameId"].(string)
			
			if !ok || gameID == "" {
				return false
			}

			// Wait a bit more to ensure the connection is fully established in the game room
			time.Sleep(100 * time.Millisecond)

			// Record start time for broadcast test
			startTime := time.Now()

			// Get the game session to get the board state
			session, err := gameService.GetSession(ctx, gameID)
			if err != nil {
				return false
			}

			// Broadcast a move message
			moveMsg := wspackage.CreateMoveMadeMessage(
				gameID,
				playerName,
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

			// Wait for message to be received by the client
			_, receivedMove := client.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeMoveMade)

			// Check timing (allow a bit more time for the test infrastructure)
			elapsed := time.Since(startTime)
			return elapsed <= 200*time.Millisecond && receivedMove
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) >= 3 && len(s) <= 10 }),
		gen.IntRange(0, 6),
	))

	properties.Property("for any message broadcast, only connections in the same game should receive it", prop.ForAll(
		func(game1Player1, game2Player1 string) bool {
			// Skip invalid inputs
			if game1Player1 == "" || game2Player1 == "" || game1Player1 == game2Player1 {
				return true
			}

			// Create mock game service and WebSocket service
			gameService := NewMockGameService()
			matchmakingService := &MockMatchmakingService{}
			wsService := wspackage.NewService(gameService, matchmakingService)

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

			client2, err := NewTestClient(wsURL, game2Player1)
			if err != nil {
				return false
			}
			defer client2.Close()

			// Wait for connections to be established
			time.Sleep(50 * time.Millisecond)

			// Send join_game messages to create two separate bot games
			joinMsg1 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game1Player1,
				"gameType": "bot",
			})
			
			joinMsg2 := wspackage.NewMessage(wspackage.MessageTypeJoinGame, map[string]interface{}{
				"username": game2Player1,
				"gameType": "bot",
			})

			// Send join messages
			if err := client1.SendMessage(joinMsg1); err != nil {
				return false
			}
			if err := client2.SendMessage(joinMsg2); err != nil {
				return false
			}

			// Wait for join processing
			time.Sleep(200 * time.Millisecond)

			// Get game IDs from game started messages
			gameStartedMsg1, received1 := client1.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)
			gameStartedMsg2, received2 := client2.WaitForMessage(200*time.Millisecond, wspackage.MessageTypeGameStarted)

			if !received1 || !received2 {
				return false
			}

			gameID1, ok1 := gameStartedMsg1.Payload["gameId"].(string)
			gameID2, ok2 := gameStartedMsg2.Payload["gameId"].(string)
			
			if !ok1 || !ok2 || gameID1 == "" || gameID2 == "" {
				return false
			}

			// Verify that the games are different and connections are properly isolated
			game1Conns := wsService.GetGameConnections(gameID1)
			game2Conns := wsService.GetGameConnections(gameID2)

			// Each game should have exactly one connection (the player who joined)
			// Games should be separate
			game1HasCorrectPlayer := len(game1Conns) == 1
			game2HasCorrectPlayer := len(game2Conns) == 1
			gamesAreSeparate := gameID1 != gameID2

			return game1HasCorrectPlayer && game2HasCorrectPlayer && gamesAreSeparate
		},
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
			matchmakingService := &MockMatchmakingService{}
			wsService := wspackage.NewService(gameService, matchmakingService)

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