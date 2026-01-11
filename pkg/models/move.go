package models

import (
	"time"

	"gorm.io/gorm"
)

// Move represents a single move in a Connect 4 game
type Move struct {
	ID        string      `json:"id" gorm:"primaryKey" validate:"required"`
	GameID    string      `json:"gameId" gorm:"not null;index" validate:"required"`
	Player    PlayerColor `json:"player" gorm:"type:varchar(10);not null" validate:"required"`
	Column    int         `json:"column" gorm:"not null" validate:"required,min=0,max=6"`
	Row       int         `json:"row" gorm:"not null" validate:"required,min=0,max=5"`
	Timestamp time.Time   `json:"timestamp" gorm:"autoCreateTime"`
	CreatedAt time.Time   `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (Move) TableName() string {
	return "moves"
}

// BeforeCreate is a GORM hook that runs before creating a move
func (m *Move) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = generateUUID()
	}
	return nil
}

// IsValid validates the move
func (m *Move) IsValid() bool {
	return m.Column >= 0 && m.Column < 7 && 
		   m.Row >= 0 && m.Row < 6 &&
		   m.Player.IsValid()
}