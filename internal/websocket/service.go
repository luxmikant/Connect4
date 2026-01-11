package websocket

import (
	"context"
	"fmt"
	"log"

	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/internal/matchmaking"
)

// Service represents the WebSocket service
type Service struct {
	hub                *Hub
	messageHandler     *GameMessageHandler
	wsHandler          *WebSocketHandler
	matchmakingService matchmaking.MatchmakingService
	config             ConnectionConfig
}

// NewService creates a new WebSocket service
func NewService(gameService game.GameService, matchmakingService matchmaking.MatchmakingService) *Service {
	config := DefaultConnectionConfig()
	
	// Create hub first
	hub := NewHub(nil, config) // messageHandler will be set later
	
	// Create message handler with matchmaking service
	messageHandler := NewGameMessageHandler(gameService, matchmakingService, hub)
	
	// Set message handler in hub
	hub.messageHandler = messageHandler
	
	// Create WebSocket handler
	wsHandler := NewWebSocketHandler(hub, config)
	
	return &Service{
		hub:                hub,
		messageHandler:     messageHandler,
		wsHandler:          wsHandler,
		matchmakingService: matchmakingService,
		config:             config,
	}
}

// Start starts the WebSocket service
func (s *Service) Start(ctx context.Context) error {
	log.Println("Starting WebSocket service...")
	
	// Start matchmaking service
	if err := s.matchmakingService.StartMatchmaking(ctx); err != nil {
		return fmt.Errorf("failed to start matchmaking service: %w", err)
	}
	
	// Start the hub in a goroutine
	go s.hub.Run()
	
	log.Println("WebSocket service started successfully")
	return nil
}

// Stop stops the WebSocket service
func (s *Service) Stop() error {
	log.Println("Stopping WebSocket service...")
	
	// Stop matchmaking service
	s.matchmakingService.StopMatchmaking()
	
	s.hub.Shutdown()
	
	log.Println("WebSocket service stopped")
	return nil
}

// GetHub returns the WebSocket hub
func (s *Service) GetHub() *Hub {
	return s.hub
}

// GetWebSocketHandler returns the WebSocket HTTP handler
func (s *Service) GetWebSocketHandler() *WebSocketHandler {
	return s.wsHandler
}

// GetConnectionCount returns the number of active connections
func (s *Service) GetConnectionCount() int {
	return s.hub.GetConnectionCount()
}

// GetActiveGames returns a list of active game IDs
func (s *Service) GetActiveGames() []string {
	return s.hub.GetActiveGames()
}

// BroadcastToGame broadcasts a message to all connections in a game
func (s *Service) BroadcastToGame(gameID string, message []byte, excludeUserID string) {
	s.hub.BroadcastToGame(gameID, message, excludeUserID)
}

// GetGameConnections returns all connections for a specific game
func (s *Service) GetGameConnections(gameID string) map[string]*Connection {
	return s.hub.GetGameConnections(gameID)
}

// IsUserConnected checks if a user is currently connected
func (s *Service) IsUserConnected(userID string) bool {
	_, exists := s.hub.GetConnection(userID)
	return exists
}

// SendMessageToUser sends a message to a specific user
func (s *Service) SendMessageToUser(userID string, message []byte) error {
	conn, exists := s.hub.GetConnection(userID)
	if !exists {
		return ErrUserNotConnected
	}
	
	conn.SendMessage(message)
	return nil
}

// NotifyGameStateChange notifies all players in a game about a state change
func (s *Service) NotifyGameStateChange(gameID string, gameState interface{}) error {
	// This is a placeholder - in practice, the actual game state should be passed
	// and properly extracted. For now, we'll use the sendGameState method from handler
	// which has access to the game service.
	return nil
}

// NotifyPlayerMove notifies all players about a move
func (s *Service) NotifyPlayerMove(gameID, player string, column, row int, board interface{}, nextTurn string, moveCount int) error {
	msg := CreateMoveMadeMessage(gameID, player, column, row, board, nextTurn, moveCount)
	
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	
	s.BroadcastToGame(gameID, data, "")
	return nil
}

// NotifyGameEnd notifies all players that a game has ended
func (s *Service) NotifyGameEnd(gameID string, winner *string, reason string, duration int) error {
	msg := CreateGameEndedMessage(gameID, winner, reason, duration)
	
	data, err := msg.ToJSON()
	if err != nil {
		return err
	}
	
	s.BroadcastToGame(gameID, data, "")
	return nil
}