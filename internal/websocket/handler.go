package websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/pkg/models"
)

// GameMessageHandler handles WebSocket messages related to game operations
type GameMessageHandler struct {
	gameService game.GameService
	hub         *Hub
}

// NewGameMessageHandler creates a new game message handler
func NewGameMessageHandler(gameService game.GameService, hub *Hub) *GameMessageHandler {
	return &GameMessageHandler{
		gameService: gameService,
		hub:         hub,
	}
}

// HandleMessage processes incoming WebSocket messages
func (h *GameMessageHandler) HandleMessage(ctx context.Context, conn *Connection, message *Message) error {
	switch message.Type {
	case MessageTypeJoinGame:
		return h.handleJoinGame(ctx, conn, message)
	case MessageTypeMakeMove:
		return h.handleMakeMove(ctx, conn, message)
	case MessageTypeReconnect:
		return h.handleReconnect(ctx, conn, message)
	case MessageTypeLeaveGame:
		return h.handleLeaveGame(ctx, conn, message)
	case MessageTypePing:
		return h.handlePing(ctx, conn, message)
	default:
		return fmt.Errorf("unknown message type: %s", message.Type)
	}
}

// handleJoinGame processes join game requests
func (h *GameMessageHandler) handleJoinGame(ctx context.Context, conn *Connection, message *Message) error {
	username, ok := message.Payload["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("invalid username")
	}

	gameType, _ := message.Payload["gameType"].(string)
	if gameType == "" {
		gameType = "pvp" // default to player vs player
	}

	log.Printf("Player %s joining game (type: %s)", username, gameType)

	// For now, create a simple bot game
	// TODO: Implement proper matchmaking in Task 8
	var session *models.GameSession
	var err error

	if gameType == "bot" {
		// Create game with bot
		session, err = h.gameService.CreateSession(ctx, username, "Bot")
	} else {
		// For PvP, we'll need matchmaking (Task 8)
		// For now, create a bot game as fallback
		session, err = h.gameService.CreateSession(ctx, username, "Bot")
	}

	if err != nil {
		return fmt.Errorf("failed to create game session: %w", err)
	}

	// Update connection with game ID
	conn.SetGameID(session.ID)

	// Add connection to game room
	h.hub.mu.Lock()
	h.hub.addToGameRoom(conn)
	h.hub.mu.Unlock()

	// Send game started message
	isBot := session.Player2 == "Bot"
	yourColor := string(session.GetPlayerColor(username))
	currentTurn := string(session.CurrentTurn)

	gameStartedMsg := CreateGameStartedMessage(
		session.ID,
		session.Player2,
		yourColor,
		currentTurn,
		isBot,
	)

	data, err := gameStartedMsg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize game started message: %w", err)
	}

	conn.SendMessage(data)

	// Send initial game state
	return h.sendGameState(ctx, conn, session.ID)
}

// handleMakeMove processes move requests
func (h *GameMessageHandler) handleMakeMove(ctx context.Context, conn *Connection, message *Message) error {
	gameID, ok := message.Payload["gameId"].(string)
	if !ok || gameID == "" {
		return fmt.Errorf("invalid game ID")
	}

	columnFloat, ok := message.Payload["column"].(float64)
	if !ok {
		return fmt.Errorf("invalid column")
	}
	column := int(columnFloat)

	username := conn.GetUserID()

	log.Printf("Player %s making move in game %s, column %d", username, gameID, column)

	// Get current game session
	session, err := h.gameService.GetSession(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game session: %w", err)
	}

	// Validate it's the player's turn
	if session.GetCurrentPlayer() != username {
		return fmt.Errorf("not your turn")
	}

	// Validate move
	if !session.Board.IsValidMove(column) {
		return fmt.Errorf("invalid move: column is full or out of bounds")
	}

	// Make the move
	playerColor := session.GetPlayerColor(username)
	row := session.Board.Height[column] // Get row before making move
	
	if err := session.Board.MakeMove(column, playerColor); err != nil {
		return fmt.Errorf("failed to make move: %w", err)
	}

	// Check for win or draw
	winner := session.Board.CheckWin()
	if winner != nil {
		// Game won
		if err := h.gameService.CompleteGame(ctx, gameID, winner); err != nil {
			return fmt.Errorf("failed to complete game: %w", err)
		}
	} else if session.Board.IsFull() {
		// Game is a draw
		if err := h.gameService.CompleteGame(ctx, gameID, nil); err != nil {
			return fmt.Errorf("failed to complete game: %w", err)
		}
	} else {
		// Switch turn and update session
		if err := h.gameService.SwitchTurn(ctx, gameID); err != nil {
			return fmt.Errorf("failed to switch turn: %w", err)
		}
	}

	// Get updated session
	updatedSession, err := h.gameService.GetSession(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get updated session: %w", err)
	}

	// Calculate move count
	moveCount := 0
	for col := 0; col < 7; col++ {
		moveCount += updatedSession.Board.Height[col]
	}

	// Broadcast move to all players in the game
	nextTurn := string(updatedSession.CurrentTurn)
	if updatedSession.Status == models.StatusCompleted {
		nextTurn = ""
	}

	moveMadeMsg := CreateMoveMadeMessage(
		gameID,
		username,
		column,
		row,
		updatedSession.Board,
		nextTurn,
		moveCount,
	)

	data, err := moveMadeMsg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize move made message: %w", err)
	}

	h.hub.BroadcastToGame(gameID, data, "")

	// If game ended, send game ended message
	if updatedSession.Status == models.StatusCompleted {
		var winnerUsername *string
		if updatedSession.Winner != nil {
			if *updatedSession.Winner == models.PlayerColorRed {
				winnerUsername = &updatedSession.Player1
			} else {
				winnerUsername = &updatedSession.Player2
			}
		}

		reason := "connect_four"
		if updatedSession.Board.IsFull() && updatedSession.Winner == nil {
			reason = "draw"
		}

		duration := int(time.Since(updatedSession.StartTime).Seconds())

		gameEndedMsg := CreateGameEndedMessage(gameID, winnerUsername, reason, duration)
		endData, err := gameEndedMsg.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize game ended message: %w", err)
		}

		h.hub.BroadcastToGame(gameID, endData, "")
	}

	return nil
}

// handleReconnect processes reconnection requests
func (h *GameMessageHandler) handleReconnect(ctx context.Context, conn *Connection, message *Message) error {
	gameID, ok := message.Payload["gameId"].(string)
	if !ok || gameID == "" {
		return fmt.Errorf("invalid game ID")
	}

	username, ok := message.Payload["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("invalid username")
	}

	log.Printf("Player %s reconnecting to game %s", username, gameID)

	// Verify the game exists and user is part of it
	session, err := h.gameService.GetSession(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game session: %w", err)
	}

	if session.Player1 != username && session.Player2 != username {
		return fmt.Errorf("user not part of this game")
	}

	// Update connection with game ID
	conn.SetGameID(gameID)

	// Add connection to game room
	h.hub.mu.Lock()
	h.hub.addToGameRoom(conn)
	h.hub.mu.Unlock()

	// Send current game state
	return h.sendGameState(ctx, conn, gameID)
}

// handleLeaveGame processes leave game requests
func (h *GameMessageHandler) handleLeaveGame(ctx context.Context, conn *Connection, message *Message) error {
	gameID := conn.GetGameID()
	username := conn.GetUserID()

	if gameID == "" {
		return fmt.Errorf("not in a game")
	}

	log.Printf("Player %s leaving game %s", username, gameID)

	// Remove from game room
	h.hub.mu.Lock()
	h.hub.removeFromGameRoom(conn)
	h.hub.mu.Unlock()

	// Clear game ID from connection
	conn.SetGameID("")

	// Notify other players
	playerLeftMsg := NewMessage(MessageTypePlayerLeft, map[string]interface{}{
		"gameId":   gameID,
		"username": username,
		"reason":   "left",
	})

	data, err := playerLeftMsg.ToJSON()
	if err == nil {
		h.hub.BroadcastToGame(gameID, data, username)
	}

	return nil
}

// handlePing processes ping requests
func (h *GameMessageHandler) handlePing(ctx context.Context, conn *Connection, message *Message) error {
	pongMsg := CreatePongMessage()
	data, err := pongMsg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize pong message: %w", err)
	}

	conn.SendMessage(data)
	return nil
}

// sendGameState sends the current game state to a connection
func (h *GameMessageHandler) sendGameState(ctx context.Context, conn *Connection, gameID string) error {
	session, err := h.gameService.GetSession(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game session: %w", err)
	}

	// Calculate move count
	moveCount := 0
	for col := 0; col < 7; col++ {
		moveCount += session.Board.Height[col]
	}

	var winnerUsername *string
	if session.Winner != nil {
		if *session.Winner == models.PlayerColorRed {
			winnerUsername = &session.Player1
		} else {
			winnerUsername = &session.Player2
		}
	}

	gameStateMsg := CreateGameStateMessage(
		session.ID,
		session.Player1,
		session.Player2,
		session.Board,
		string(session.CurrentTurn),
		string(session.Status),
		winnerUsername,
		moveCount,
		session.StartTime,
	)

	data, err := gameStateMsg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize game state message: %w", err)
	}

	conn.SendMessage(data)
	return nil
}

// WebSocketHandler handles WebSocket upgrade requests
type WebSocketHandler struct {
	hub    *Hub
	config ConnectionConfig
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, config ConnectionConfig) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		config: config,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from query parameter or header
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Optional game ID for reconnection
	gameID := c.Query("gameId")

	// Upgrade HTTP connection to WebSocket
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	// Create connection wrapper
	wsConn := NewConnection(conn, userID, gameID, h.hub)

	// Register connection with hub
	h.hub.RegisterConnection(wsConn)

	// Start connection pumps
	wsConn.Start(c.Request.Context(), h.config)

	log.Printf("WebSocket connection established: user=%s, game=%s", userID, gameID)
}