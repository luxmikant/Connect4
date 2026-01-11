package websocket

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Client to Server messages
	MessageTypeJoinQueue   MessageType = "join_queue"    // New: Join matchmaking queue
	MessageTypeLeaveQueue  MessageType = "leave_queue"   // New: Leave matchmaking queue
	MessageTypeJoinGame    MessageType = "join_game"
	MessageTypeMakeMove    MessageType = "make_move"
	MessageTypeReconnect   MessageType = "reconnect"
	MessageTypeLeaveGame   MessageType = "leave_game"
	MessageTypePing        MessageType = "ping"

	// Server to Client messages
	MessageTypeQueueJoined    MessageType = "queue_joined"    // New: Joined matchmaking queue
	MessageTypeQueueStatus    MessageType = "queue_status"    // New: Queue status update
	MessageTypeMatchFound     MessageType = "match_found"     // New: Match found notification
	MessageTypeGameStarted    MessageType = "game_started"
	MessageTypeMoveMade       MessageType = "move_made"
	MessageTypeGameEnded      MessageType = "game_ended"
	MessageTypeGameState      MessageType = "game_state"
	MessageTypePlayerJoined   MessageType = "player_joined"
	MessageTypePlayerLeft     MessageType = "player_left"
	MessageTypeError          MessageType = "error"
	MessageTypePong           MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(msgType MessageType, payload map[string]interface{}) *Message {
	return &Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// ToJSON converts the message to JSON bytes
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates a message from JSON bytes
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// JoinQueuePayload represents the payload for joining matchmaking queue
type JoinQueuePayload struct {
	Username string `json:"username"`
}

// QueueJoinedPayload represents the payload when successfully joined queue
type QueueJoinedPayload struct {
	Position      int    `json:"position"`
	EstimatedWait string `json:"estimatedWait"`
}

// QueueStatusPayload represents the current queue status
type QueueStatusPayload struct {
	InQueue       bool   `json:"inQueue"`
	Position      int    `json:"position"`
	WaitTime      string `json:"waitTime"`
	TimeRemaining string `json:"timeRemaining"`
}

// MatchFoundPayload represents when a match is found
type MatchFoundPayload struct {
	GameID   string `json:"gameId"`
	Opponent string `json:"opponent"`
	IsBot    bool   `json:"isBot"`
}

// JoinGamePayload represents the payload for joining a game
type JoinGamePayload struct {
	Username string `json:"username"`
	GameType string `json:"gameType,omitempty"` // "pvp" or "bot"
}

// MakeMovePayload represents the payload for making a move
type MakeMovePayload struct {
	GameID string `json:"gameId"`
	Column int    `json:"column"`
}

// ReconnectPayload represents the payload for reconnecting to a game
type ReconnectPayload struct {
	GameID   string `json:"gameId"`
	Username string `json:"username"`
}

// GameStartedPayload represents the payload when a game starts
type GameStartedPayload struct {
	GameID      string `json:"gameId"`
	Opponent    string `json:"opponent"`
	YourColor   string `json:"yourColor"`
	CurrentTurn string `json:"currentTurn"`
	IsBot       bool   `json:"isBot,omitempty"`
}

// MoveMadePayload represents the payload when a move is made
type MoveMadePayload struct {
	GameID    string      `json:"gameId"`
	Player    string      `json:"player"`
	Column    int         `json:"column"`
	Row       int         `json:"row"`
	Board     interface{} `json:"board"`
	NextTurn  string      `json:"nextTurn"`
	MoveCount int         `json:"moveCount"`
}

// GameEndedPayload represents the payload when a game ends
type GameEndedPayload struct {
	GameID   string  `json:"gameId"`
	Winner   *string `json:"winner"`
	Reason   string  `json:"reason"`
	Duration int     `json:"duration"` // in seconds
}

// GameStatePayload represents the current game state
type GameStatePayload struct {
	GameID      string      `json:"gameId"`
	Player1     string      `json:"player1"`
	Player2     string      `json:"player2"`
	Board       interface{} `json:"board"`
	CurrentTurn string      `json:"currentTurn"`
	Status      string      `json:"status"`
	Winner      *string     `json:"winner,omitempty"`
	MoveCount   int         `json:"moveCount"`
	StartTime   time.Time   `json:"startTime"`
}

// PlayerJoinedPayload represents when a player joins
type PlayerJoinedPayload struct {
	GameID   string `json:"gameId"`
	Username string `json:"username"`
	Color    string `json:"color"`
}

// PlayerLeftPayload represents when a player leaves
type PlayerLeftPayload struct {
	GameID   string `json:"gameId"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Helper functions to create specific message types

// CreateJoinQueueMessage creates a join queue message
func CreateJoinQueueMessage(username string) *Message {
	return NewMessage(MessageTypeJoinQueue, map[string]interface{}{
		"username": username,
	})
}

// CreateLeaveQueueMessage creates a leave queue message
func CreateLeaveQueueMessage() *Message {
	return NewMessage(MessageTypeLeaveQueue, map[string]interface{}{})
}

// CreateQueueJoinedMessage creates a queue joined message
func CreateQueueJoinedMessage(position int, estimatedWait string) *Message {
	return NewMessage(MessageTypeQueueJoined, map[string]interface{}{
		"position":      position,
		"estimatedWait": estimatedWait,
	})
}

// CreateQueueStatusMessage creates a queue status message
func CreateQueueStatusMessage(inQueue bool, position int, waitTime, timeRemaining string) *Message {
	return NewMessage(MessageTypeQueueStatus, map[string]interface{}{
		"inQueue":       inQueue,
		"position":      position,
		"waitTime":      waitTime,
		"timeRemaining": timeRemaining,
	})
}

// CreateMatchFoundMessage creates a match found message
func CreateMatchFoundMessage(gameID, opponent string, isBot bool) *Message {
	return NewMessage(MessageTypeMatchFound, map[string]interface{}{
		"gameId":   gameID,
		"opponent": opponent,
		"isBot":    isBot,
	})
}

// CreateJoinGameMessage creates a join game message
func CreateJoinGameMessage(username, gameType string) *Message {
	return NewMessage(MessageTypeJoinGame, map[string]interface{}{
		"username": username,
		"gameType": gameType,
	})
}

// CreateMakeMoveMessage creates a make move message
func CreateMakeMoveMessage(gameID string, column int) *Message {
	return NewMessage(MessageTypeMakeMove, map[string]interface{}{
		"gameId": gameID,
		"column": column,
	})
}

// CreateReconnectMessage creates a reconnect message
func CreateReconnectMessage(gameID, username string) *Message {
	return NewMessage(MessageTypeReconnect, map[string]interface{}{
		"gameId":   gameID,
		"username": username,
	})
}

// CreateGameStartedMessage creates a game started message
func CreateGameStartedMessage(gameID, opponent, yourColor, currentTurn string, isBot bool) *Message {
	return NewMessage(MessageTypeGameStarted, map[string]interface{}{
		"gameId":      gameID,
		"opponent":    opponent,
		"yourColor":   yourColor,
		"currentTurn": currentTurn,
		"isBot":       isBot,
	})
}

// CreateMoveMadeMessage creates a move made message
func CreateMoveMadeMessage(gameID, player string, column, row int, board interface{}, nextTurn string, moveCount int) *Message {
	return NewMessage(MessageTypeMoveMade, map[string]interface{}{
		"gameId":    gameID,
		"player":    player,
		"column":    column,
		"row":       row,
		"board":     board,
		"nextTurn":  nextTurn,
		"moveCount": moveCount,
	})
}

// CreateGameEndedMessage creates a game ended message
func CreateGameEndedMessage(gameID string, winner *string, reason string, duration int) *Message {
	return NewMessage(MessageTypeGameEnded, map[string]interface{}{
		"gameId":   gameID,
		"winner":   winner,
		"reason":   reason,
		"duration": duration,
	})
}

// CreateGameStateMessage creates a game state message
func CreateGameStateMessage(gameID, player1, player2 string, board interface{}, currentTurn, status string, winner *string, moveCount int, startTime time.Time) *Message {
	return NewMessage(MessageTypeGameState, map[string]interface{}{
		"gameId":      gameID,
		"player1":     player1,
		"player2":     player2,
		"board":       board,
		"currentTurn": currentTurn,
		"status":      status,
		"winner":      winner,
		"moveCount":   moveCount,
		"startTime":   startTime,
	})
}

// CreateErrorMessage creates an error message
func CreateErrorMessage(code, message, details string) *Message {
	return NewMessage(MessageTypeError, map[string]interface{}{
		"code":    code,
		"message": message,
		"details": details,
	})
}

// CreatePongMessage creates a pong message
func CreatePongMessage() *Message {
	return NewMessage(MessageTypePong, map[string]interface{}{})
}