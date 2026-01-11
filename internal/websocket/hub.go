package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active connections and broadcasts messages to the connections
type Hub struct {
	// Registered connections mapped by user ID
	connections map[string]*Connection

	// Game rooms mapped by game ID
	gameRooms map[string]map[string]*Connection

	// Register requests from connections
	register chan *Connection

	// Unregister requests from connections
	unregister chan *Connection

	// Broadcast messages to all connections in a game
	broadcast chan *BroadcastMessage

	// Message handler for processing incoming messages
	messageHandler MessageHandler

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	config ConnectionConfig
}

// BroadcastMessage represents a message to be broadcast to a game room
type BroadcastMessage struct {
	GameID  string
	Message []byte
	Exclude string // Optional: exclude this user ID from broadcast
}

// MessageHandler interface for handling incoming WebSocket messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, conn *Connection, message *Message) error
}

// NewHub creates a new WebSocket hub
func NewHub(messageHandler MessageHandler, config ConnectionConfig) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		connections:    make(map[string]*Connection),
		gameRooms:      make(map[string]map[string]*Connection),
		register:       make(chan *Connection),
		unregister:     make(chan *Connection),
		broadcast:      make(chan *BroadcastMessage),
		messageHandler: messageHandler,
		ctx:            ctx,
		cancel:         cancel,
		config:         config,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Start cleanup routine
	go h.cleanupRoutine()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case message := <-h.broadcast:
			h.broadcastToGame(message)

		case <-h.ctx.Done():
			return
		}
	}
}

// RegisterConnection registers a new connection
func (h *Hub) RegisterConnection(conn *Connection) {
	h.register <- conn
}

// UnregisterConnection unregisters a connection
func (h *Hub) UnregisterConnection(conn *Connection) {
	h.unregister <- conn
}

// BroadcastToGame broadcasts a message to all connections in a game
func (h *Hub) BroadcastToGame(gameID string, message []byte, excludeUserID string) {
	h.broadcast <- &BroadcastMessage{
		GameID:  gameID,
		Message: message,
		Exclude: excludeUserID,
	}
}

// GetConnection returns a connection by user ID
func (h *Hub) GetConnection(userID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, exists := h.connections[userID]
	return conn, exists
}

// UpdateConnectionUserID updates a connection's user ID in the hub
// This should be called when the user authenticates with a username after connecting
func (h *Hub) UpdateConnectionUserID(conn *Connection, oldUserID, newUserID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from old mapping if exists
	if oldUserID != "" {
		if existingConn, exists := h.connections[oldUserID]; exists && existingConn == conn {
			delete(h.connections, oldUserID)
		}
	}

	// Check if new userID already has a connection
	if existingConn, exists := h.connections[newUserID]; exists && existingConn != conn {
		// Close existing connection for this user
		existingConn.Close()
		h.removeFromGameRoom(existingConn)
	}

	// Register under new userID
	h.connections[newUserID] = conn

	log.Printf("Connection userID updated: %s -> %s, total_connections=%d",
		oldUserID, newUserID, len(h.connections))
}

// GetGameConnections returns all connections for a specific game
func (h *Hub) GetGameConnections(gameID string) map[string]*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	gameConns := make(map[string]*Connection)
	if room, exists := h.gameRooms[gameID]; exists {
		for userID, conn := range room {
			gameConns[userID] = conn
		}
	}
	return gameConns
}

// GetActiveGames returns a list of active game IDs
func (h *Hub) GetActiveGames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	games := make([]string, 0, len(h.gameRooms))
	for gameID := range h.gameRooms {
		games = append(games, gameID)
	}
	return games
}

// GetConnectionCount returns the total number of active connections
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	h.cancel()

	// Close all connections
	h.mu.Lock()
	for _, conn := range h.connections {
		conn.Close()
	}
	h.mu.Unlock()
}

// registerConnection handles registering a new connection
func (h *Hub) registerConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userID := conn.GetUserID()
	gameID := conn.GetGameID()

	// Check if user already has a connection
	if existingConn, exists := h.connections[userID]; exists {
		// Close existing connection
		existingConn.Close()
		h.removeFromGameRoom(existingConn)
	}

	// Register new connection
	h.connections[userID] = conn

	// Add to game room if gameID is provided
	if gameID != "" {
		h.addToGameRoom(conn)
	}

	log.Printf("Connection registered: user=%s, game=%s, total_connections=%d",
		userID, gameID, len(h.connections))
}

// unregisterConnection handles unregistering a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userID := conn.GetUserID()

	// Remove from connections map
	if _, exists := h.connections[userID]; exists {
		delete(h.connections, userID)
		conn.Close()
	}

	// Remove from game room
	h.removeFromGameRoom(conn)

	log.Printf("Connection unregistered: user=%s, total_connections=%d",
		userID, len(h.connections))
}

// addToGameRoom adds a connection to a game room
func (h *Hub) addToGameRoom(conn *Connection) {
	gameID := conn.GetGameID()
	userID := conn.GetUserID()

	if gameID == "" {
		return
	}

	if h.gameRooms[gameID] == nil {
		h.gameRooms[gameID] = make(map[string]*Connection)
	}

	h.gameRooms[gameID][userID] = conn

	log.Printf("Added to game room: user=%s, game=%s, room_size=%d",
		userID, gameID, len(h.gameRooms[gameID]))
}

// removeFromGameRoom removes a connection from its game room
func (h *Hub) removeFromGameRoom(conn *Connection) {
	gameID := conn.GetGameID()
	userID := conn.GetUserID()

	if gameID == "" {
		return
	}

	if room, exists := h.gameRooms[gameID]; exists {
		delete(room, userID)

		// Remove empty game rooms
		if len(room) == 0 {
			delete(h.gameRooms, gameID)
			log.Printf("Removed empty game room: game=%s", gameID)
		} else {
			log.Printf("Removed from game room: user=%s, game=%s, room_size=%d",
				userID, gameID, len(room))
		}
	}
}

// broadcastToGame broadcasts a message to all connections in a game room
func (h *Hub) broadcastToGame(msg *BroadcastMessage) {
	h.mu.RLock()
	room, exists := h.gameRooms[msg.GameID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// Create a copy of connections to avoid holding lock during send
	connections := make([]*Connection, 0, len(room))
	for userID, conn := range room {
		if userID != msg.Exclude {
			connections = append(connections, conn)
		}
	}
	h.mu.RUnlock()

	// Send message to all connections
	for _, conn := range connections {
		conn.SendMessage(msg.Message)
	}

	log.Printf("Broadcast to game: game=%s, recipients=%d, excluded=%s",
		msg.GameID, len(connections), msg.Exclude)
}

// handleMessage processes incoming messages from connections
func (h *Hub) handleMessage(conn *Connection, message *Message) {
	if h.messageHandler != nil {
		if err := h.messageHandler.HandleMessage(h.ctx, conn, message); err != nil {
			log.Printf("Error handling message: %v", err)

			// Send error response to client
			errorMsg := &Message{
				Type: MessageTypeError,
				Payload: map[string]interface{}{
					"error": err.Error(),
				},
			}

			if data, err := json.Marshal(errorMsg); err == nil {
				conn.SendMessage(data)
			}
		}
	}
}

// cleanupRoutine periodically cleans up stale connections
func (h *Hub) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.cleanupStaleConnections()
		case <-h.ctx.Done():
			return
		}
	}
}

// cleanupStaleConnections removes connections that haven't been seen recently
func (h *Hub) cleanupStaleConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	staleThreshold := time.Now().Add(-5 * time.Minute)
	staleConnections := make([]*Connection, 0)

	for userID, conn := range h.connections {
		if conn.GetLastSeen().Before(staleThreshold) {
			staleConnections = append(staleConnections, conn)
			delete(h.connections, userID)
		}
	}

	// Clean up stale connections outside the lock
	go func() {
		for _, conn := range staleConnections {
			h.removeFromGameRoom(conn)
			conn.Close()
		}

		if len(staleConnections) > 0 {
			log.Printf("Cleaned up %d stale connections", len(staleConnections))
		}
	}()
}
