package models

import (
	"time"

	"gorm.io/gorm"
)

// PlayerStats represents player statistics
type PlayerStats struct {
	ID          string    `json:"id" gorm:"primaryKey" validate:"required"`
	Username    string    `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=20"`
	GamesPlayed int       `json:"gamesPlayed" gorm:"default:0" validate:"min=0"`
	GamesWon    int       `json:"gamesWon" gorm:"default:0" validate:"min=0"`
	WinRate     float64   `json:"winRate" gorm:"default:0.0" validate:"min=0,max=1"`
	AvgGameTime int       `json:"avgGameTime" gorm:"default:0" validate:"min=0"` // In seconds
	LastPlayed  time.Time `json:"lastPlayed"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (PlayerStats) TableName() string {
	return "player_stats"
}

// BeforeCreate is a GORM hook that runs before creating player stats
func (ps *PlayerStats) BeforeCreate(tx *gorm.DB) error {
	if ps.ID == "" {
		ps.ID = generateUUID()
	}
	return nil
}

// CalculateWinRate calculates and updates the win rate
func (ps *PlayerStats) CalculateWinRate() {
	if ps.GamesPlayed > 0 {
		ps.WinRate = float64(ps.GamesWon) / float64(ps.GamesPlayed)
	} else {
		ps.WinRate = 0.0
	}
}

// UpdateGameStats updates statistics after a game
func (ps *PlayerStats) UpdateGameStats(won bool, gameDuration int) {
	ps.GamesPlayed++
	if won {
		ps.GamesWon++
	}

	// Update average game time
	if ps.AvgGameTime == 0 {
		ps.AvgGameTime = gameDuration
	} else {
		ps.AvgGameTime = (ps.AvgGameTime*(ps.GamesPlayed-1) + gameDuration) / ps.GamesPlayed
	}

	ps.CalculateWinRate()
	ps.LastPlayed = time.Now()
}

// AnalyticsSnapshot represents a point-in-time snapshot of game analytics metrics
// Used by the analytics service to persist aggregated metrics (Requirement 10.5)
type AnalyticsSnapshot struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	Timestamp          time.Time `json:"timestamp" gorm:"autoCreateTime;index"`
	GamesCompletedHour int64     `json:"gamesCompletedHour" gorm:"default:0"`
	GamesCompletedDay  int64     `json:"gamesCompletedDay" gorm:"default:0"`
	AvgGameDurationSec int64     `json:"avgGameDurationSec" gorm:"default:0"`
	MinGameDurationSec int64     `json:"minGameDurationSec" gorm:"default:0"`
	MaxGameDurationSec int64     `json:"maxGameDurationSec" gorm:"default:0"`
	TotalMoves         int64     `json:"totalMoves" gorm:"default:0"`
	UniquePlayersHour  int64     `json:"uniquePlayersHour" gorm:"default:0"`
	ActiveGames        int64     `json:"activeGames" gorm:"default:0"`
	CreatedAt          time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (AnalyticsSnapshot) TableName() string {
	return "analytics_snapshots"
}

// BeforeCreate is a GORM hook that runs before creating an analytics snapshot
func (as *AnalyticsSnapshot) BeforeCreate(tx *gorm.DB) error {
	if as.ID == "" {
		as.ID = generateUUID()
	}
	return nil
}
