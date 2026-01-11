package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents a WebSocket connection with metadata
type Connection struct {
	conn     *websocket.Conn
	userID   string
	gameID   string
	send     chan []byte
	hub      *Hub
	mu       sync.RWMutex
	lastSeen time.Time
}

// ConnectionConfig holds configuration for WebSocket connections
type ConnectionConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// DefaultConnectionConfig returns default configuration
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     54 * time.Second, // Must be less than PongWait
		MaxMessageSize: 512,
	}
}

// NewConnection creates a new WebSocket connection
func NewConnection(conn *websocket.Conn, userID, gameID string, hub *Hub) *Connection {
	return &Connection{
		conn:     conn,
		userID:   userID,
		gameID:   gameID,
		send:     make(chan []byte, 256),
		hub:      hub,
		lastSeen: time.Now(),
	}
}

// GetUserID returns the user ID associated with this connection
func (c *Connection) GetUserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// GetGameID returns the game ID associated with this connection
func (c *Connection) GetGameID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameID
}

// SetGameID sets the game ID for this connection
func (c *Connection) SetGameID(gameID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gameID = gameID
}

// UpdateLastSeen updates the last seen timestamp
func (c *Connection) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSeen = time.Now()
}

// GetLastSeen returns the last seen timestamp
func (c *Connection) GetLastSeen() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastSeen
}

// SendMessage sends a message to the WebSocket connection
func (c *Connection) SendMessage(message []byte) {
	select {
	case c.send <- message:
	default:
		// Channel is full, close connection
		c.hub.unregister <- c
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *Connection) readPump(config ConnectionConfig) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(config.PongWait))
		c.UpdateLastSeen()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.UpdateLastSeen()
		
		// Parse and handle the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Send message to hub for processing
		c.hub.handleMessage(c, &msg)
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Connection) writePump(config ConnectionConfig) {
	ticker := time.NewTicker(config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start begins the connection's read and write pumps
func (c *Connection) Start(ctx context.Context, config ConnectionConfig) {
	go c.writePump(config)
	go c.readPump(config)
}

// Close closes the connection
func (c *Connection) Close() {
	close(c.send)
}

// Upgrader configures the WebSocket upgrader
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}