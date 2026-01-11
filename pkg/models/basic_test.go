package models_test

import (
	"testing"
	"time"

	"connect4-multiplayer/pkg/models"
)

func TestBasicModelFunctionality(t *testing.T) {
	// Test Player creation
	player := &models.Player{
		ID:       "test-player-1",
		Username: "testuser",
	}
	
	if player.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", player.Username)
	}

	// Test GameSession creation
	gameSession := &models.GameSession{
		ID:          "test-game-1",
		Player1:     "player1",
		Player2:     "player2",
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusWaiting,
	}
	gameSession.Board = models.NewBoard()

	if !gameSession.Board.IsValidMove(0) {
		t.Error("Expected column 0 to be a valid move on new board")
	}

	// Test Move validation
	move := &models.Move{
		ID:     "test-move-1",
		GameID: "test-game-1",
		Player: models.PlayerColorRed,
		Column: 3,
		Row:    0,
	}

	if !move.IsValid() {
		t.Error("Expected move to be valid")
	}

	// Test PlayerStats calculation
	stats := &models.PlayerStats{
		ID:          "test-stats-1",
		Username:    "testuser",
		GamesPlayed: 10,
		GamesWon:    7,
	}
	stats.CalculateWinRate()

	expectedWinRate := 0.7
	if stats.WinRate != expectedWinRate {
		t.Errorf("Expected win rate %.2f, got %.2f", expectedWinRate, stats.WinRate)
	}

	t.Log("All basic model tests passed")
}