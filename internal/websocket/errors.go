package websocket

import "errors"

// WebSocket-specific errors
var (
	ErrUserNotConnected    = errors.New("user not connected")
	ErrGameNotFound        = errors.New("game not found")
	ErrInvalidMessage      = errors.New("invalid message format")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConnectionClosed    = errors.New("connection closed")
	ErrMessageTooLarge     = errors.New("message too large")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrInvalidGameState    = errors.New("invalid game state")
	ErrPlayerNotInGame     = errors.New("player not in game")
	ErrGameAlreadyEnded    = errors.New("game already ended")
)