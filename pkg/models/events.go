package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// EventType represents the type of game event
type EventType string

const (
	EventGameStarted    EventType = "game_started"
	EventMoveMade       EventType = "move_made"
	EventGameCompleted  EventType = "game_completed"
	EventPlayerJoined   EventType = "player_joined"
	EventPlayerLeft     EventType = "player_left"
	EventPlayerReconnected EventType = "player_reconnected"
)

// EventMetadata represents metadata for game events
type EventMetadata map[string]interface{}

// Scan implements the sql.Scanner interface for GORM
func (em *EventMetadata) Scan(value interface{}) error {
	if value == nil {
		*em = make(EventMetadata)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return ErrInvalidEventData
	}
	
	return json.Unmarshal(bytes, em)
}

// Value implements the driver.Valuer interface for GORM
func (em EventMetadata) Value() (driver.Value, error) {
	return json.Marshal(em)
}

// GameEvent represents an analytics event
type GameEvent struct {
	ID        string        `json:"id" gorm:"primaryKey" validate:"required"`
	EventType EventType     `json:"eventType" gorm:"type:varchar(50);not null;index" validate:"required"`
	GameID    string        `json:"gameId" gorm:"not null;index" validate:"required"`
	PlayerID  string        `json:"playerId" gorm:"not null" validate:"required"`
	Timestamp time.Time     `json:"timestamp" gorm:"autoCreateTime;index"`
	Metadata  EventMetadata `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time     `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (GameEvent) TableName() string {
	return "game_events"
}

// BeforeCreate is a GORM hook that runs before creating a game event
func (ge *GameEvent) BeforeCreate(tx *gorm.DB) error {
	if ge.ID == "" {
		ge.ID = generateUUID()
	}
	return nil
}

// NewGameStartedEvent creates a new game started event
func NewGameStartedEvent(gameID, player1, player2 string) *GameEvent {
	return &GameEvent{
		EventType: EventGameStarted,
		GameID:    gameID,
		PlayerID:  player1,
		Timestamp: time.Now(),
		Metadata: EventMetadata{
			"player1": player1,
			"player2": player2,
		},
	}
}

// NewMoveMadeEvent creates a new move made event
func NewMoveMadeEvent(gameID, playerID string, column, row int) *GameEvent {
	return &GameEvent{
		EventType: EventMoveMade,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: EventMetadata{
			"column": column,
			"row":    row,
		},
	}
}

// NewGameCompletedEvent creates a new game completed event
func NewGameCompletedEvent(gameID, winner, loser string, duration int) *GameEvent {
	return &GameEvent{
		EventType: EventGameCompleted,
		GameID:    gameID,
		PlayerID:  winner,
		Timestamp: time.Now(),
		Metadata: EventMetadata{
			"winner":   winner,
			"loser":    loser,
			"duration": duration,
		},
	}
}

// NewPlayerJoinedEvent creates a new player joined event
func NewPlayerJoinedEvent(gameID, playerID string) *GameEvent {
	return &GameEvent{
		EventType: EventPlayerJoined,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata:  EventMetadata{},
	}
}