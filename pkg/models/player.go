package models

import (
	"time"

	"gorm.io/gorm"
)

// Player represents a game player
type Player struct {
	ID        string    `json:"id" gorm:"primaryKey" validate:"required"`
	Username  string    `json:"username" gorm:"uniqueIndex" validate:"required,min=3,max=20"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (Player) TableName() string {
	return "players"
}

// BeforeCreate is a GORM hook that runs before creating a player
func (p *Player) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		// Generate UUID if not provided
		p.ID = generateUUID()
	}
	return nil
}

// PlayerColor represents the color of a player's discs
type PlayerColor string

const (
	PlayerColorRed    PlayerColor = "red"
	PlayerColorYellow PlayerColor = "yellow"
)

// String returns the string representation of PlayerColor
func (pc PlayerColor) String() string {
	return string(pc)
}

// IsValid checks if the player color is valid
func (pc PlayerColor) IsValid() bool {
	return pc == PlayerColorRed || pc == PlayerColorYellow
}