package models

import "errors"

// Common errors used across the application
var (
	ErrInvalidMove      = errors.New("invalid move")
	ErrGameNotFound     = errors.New("game not found")
	ErrPlayerNotFound   = errors.New("player not found")
	ErrMoveNotFound     = errors.New("move not found")
	ErrGameFull         = errors.New("game is full")
	ErrNotPlayerTurn    = errors.New("not player's turn")
	ErrGameEnded        = errors.New("game has ended")
	ErrInvalidBoardData = errors.New("invalid board data")
	ErrInvalidEventData = errors.New("invalid event data")
	ErrDuplicateUsername = errors.New("username already exists in active session")
)

// GameError represents a structured error for API responses
type GameError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (ge *GameError) Error() string {
	return ge.Message
}

// NewGameError creates a new GameError
func NewGameError(code, message, details string) *GameError {
	return &GameError{
		Code:    code,
		Message: message,
		Details: details,
	}
}