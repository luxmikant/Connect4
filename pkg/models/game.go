package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// GameStatus represents the status of a game
type GameStatus string

const (
	StatusWaiting    GameStatus = "waiting"
	StatusInProgress GameStatus = "in_progress"
	StatusCompleted  GameStatus = "completed"
	StatusAbandoned  GameStatus = "abandoned"
)

// GameSession represents a Connect 4 game session
type GameSession struct {
	ID          string       `json:"id" gorm:"primaryKey" validate:"required"`
	Player1     string       `json:"player1" gorm:"not null" validate:"required,min=3,max=20"`
	Player2     string       `json:"player2" gorm:"not null" validate:"required,min=3,max=20"`
	Board       Board        `json:"board" gorm:"type:jsonb"`
	CurrentTurn PlayerColor  `json:"currentTurn" gorm:"type:varchar(10)" validate:"required"`
	Status      GameStatus   `json:"status" gorm:"type:varchar(20);index" validate:"required"`
	Winner      *PlayerColor `json:"winner,omitempty" gorm:"type:varchar(10)"`
	StartTime   time.Time    `json:"startTime" gorm:"autoCreateTime"`
	EndTime     *time.Time   `json:"endTime,omitempty"`
	MoveHistory []Move       `json:"moveHistory" gorm:"foreignKey:GameID"`
	CreatedAt   time.Time    `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"autoUpdateTime"`
	// Custom room fields
	RoomCode  *string `json:"roomCode,omitempty" gorm:"type:varchar(8);uniqueIndex:idx_game_sessions_room_code,where:room_code IS NOT NULL"`
	IsCustom  bool    `json:"isCustom" gorm:"default:false;not null;index:idx_game_sessions_is_custom,where:is_custom = true"`
	CreatedBy *string `json:"createdBy,omitempty" gorm:"type:varchar(255)"`
}

// TableName returns the table name for GORM
func (GameSession) TableName() string {
	return "game_sessions"
}

// BeforeCreate is a GORM hook that runs before creating a game session
func (gs *GameSession) BeforeCreate(tx *gorm.DB) error {
	if gs.ID == "" {
		gs.ID = generateUUID()
	}
	if gs.Status == "" {
		gs.Status = StatusWaiting
	}
	if gs.CurrentTurn == "" {
		gs.CurrentTurn = PlayerColorRed
	}
	// Initialize empty board
	gs.Board = NewBoard()
	return nil
}

// IsActive returns true if the game is currently active
func (gs *GameSession) IsActive() bool {
	return gs.Status == StatusInProgress
}

// IsCompleted returns true if the game is completed
func (gs *GameSession) IsCompleted() bool {
	return gs.Status == StatusCompleted
}

// GetCurrentPlayer returns the username of the current player
func (gs *GameSession) GetCurrentPlayer() string {
	if gs.CurrentTurn == PlayerColorRed {
		return gs.Player1
	}
	return gs.Player2
}

// GetPlayerColor returns the color for a given player
func (gs *GameSession) GetPlayerColor(username string) PlayerColor {
	if username == gs.Player1 {
		return PlayerColorRed
	}
	return PlayerColorYellow
}

// Board represents the Connect 4 game board
type Board struct {
	Grid   [6][7]PlayerColor `json:"cells"`
	Height [7]int            `json:"height"`
}

// NewBoard creates a new empty board
func NewBoard() Board {
	return Board{
		Grid:   [6][7]PlayerColor{},
		Height: [7]int{},
	}
}

// IsValidMove checks if a move is valid
func (b *Board) IsValidMove(column int) bool {
	return column >= 0 && column < 7 && b.Height[column] < 6
}

// MakeMove makes a move on the board
func (b *Board) MakeMove(column int, player PlayerColor) error {
	if !b.IsValidMove(column) {
		return ErrInvalidMove
	}

	row := b.Height[column]
	b.Grid[row][column] = player
	b.Height[column]++

	return nil
}

// CheckWin checks if there's a winner on the board
func (b *Board) CheckWin() *PlayerColor {
	// Check horizontal
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			if b.Grid[row][col] != "" &&
				b.Grid[row][col] == b.Grid[row][col+1] &&
				b.Grid[row][col] == b.Grid[row][col+2] &&
				b.Grid[row][col] == b.Grid[row][col+3] {
				winner := b.Grid[row][col]
				return &winner
			}
		}
	}

	// Check vertical
	for row := 0; row < 3; row++ {
		for col := 0; col < 7; col++ {
			if b.Grid[row][col] != "" &&
				b.Grid[row][col] == b.Grid[row+1][col] &&
				b.Grid[row][col] == b.Grid[row+2][col] &&
				b.Grid[row][col] == b.Grid[row+3][col] {
				winner := b.Grid[row][col]
				return &winner
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			if b.Grid[row][col] != "" &&
				b.Grid[row][col] == b.Grid[row+1][col+1] &&
				b.Grid[row][col] == b.Grid[row+2][col+2] &&
				b.Grid[row][col] == b.Grid[row+3][col+3] {
				winner := b.Grid[row][col]
				return &winner
			}
		}
	}

	// Check diagonal (top-right to bottom-left)
	for row := 0; row < 3; row++ {
		for col := 3; col < 7; col++ {
			if b.Grid[row][col] != "" &&
				b.Grid[row][col] == b.Grid[row+1][col-1] &&
				b.Grid[row][col] == b.Grid[row+2][col-2] &&
				b.Grid[row][col] == b.Grid[row+3][col-3] {
				winner := b.Grid[row][col]
				return &winner
			}
		}
	}

	return nil
}

// IsFull checks if the board is full (draw condition)
func (b *Board) IsFull() bool {
	for col := 0; col < 7; col++ {
		if b.Height[col] < 6 {
			return false
		}
	}
	return true
}

// Scan implements the sql.Scanner interface for GORM
func (b *Board) Scan(value interface{}) error {
	if value == nil {
		*b = NewBoard()
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrInvalidBoardData
	}

	return json.Unmarshal(bytes, b)
}

// Value implements the driver.Valuer interface for GORM
func (b Board) Value() (driver.Value, error) {
	return json.Marshal(b)
}
