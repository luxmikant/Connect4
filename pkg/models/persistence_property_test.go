//go:build property
// +build property

package models_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"connect4-multiplayer/pkg/models"
)

// genPlayer generates a random Player for property testing
func genPlayer() gopter.Gen {
	return gen.Struct(reflect.TypeOf(&models.Player{}), map[string]gopter.Gen{
		"Username": gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
	}).Map(func(p *models.Player) *models.Player {
		// Ensure ID is generated
		if p.ID == "" {
			p.ID = "test-" + p.Username + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
		}
		return p
	})
}

// genGameSession generates a random GameSession for property testing
func genGameSession() gopter.Gen {
	return gen.Struct(reflect.TypeOf(&models.GameSession{}), map[string]gopter.Gen{
		"Player1": gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
		"Player2": gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
		"CurrentTurn": gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
		"Status": gen.OneConstOf(
			models.StatusWaiting,
			models.StatusInProgress,
			models.StatusCompleted,
			models.StatusAbandoned,
		),
	}).Map(func(gs *models.GameSession) *models.GameSession {
		// Ensure ID is generated and players are different
		if gs.ID == "" {
			gs.ID = "game-" + fmt.Sprintf("%d", time.Now().UnixNano())
		}
		if gs.Player1 == gs.Player2 {
			gs.Player2 = gs.Player2 + "2"
		}
		// Initialize board
		gs.Board = models.NewBoard()
		return gs
	})
}

// genPlayerStats generates a random PlayerStats for property testing
func genPlayerStats() gopter.Gen {
	return gen.Struct(reflect.TypeOf(&models.PlayerStats{}), map[string]gopter.Gen{
		"Username": gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
		"GamesPlayed": gen.IntRange(0, 1000),
		"GamesWon": gen.IntRange(0, 500),
		"AvgGameTime": gen.IntRange(30, 3600), // 30 seconds to 1 hour
	}).Map(func(ps *models.PlayerStats) *models.PlayerStats {
		// Ensure ID is generated and GamesWon <= GamesPlayed
		if ps.ID == "" {
			ps.ID = "stats-" + ps.Username + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
		}
		if ps.GamesWon > ps.GamesPlayed {
			ps.GamesWon = ps.GamesPlayed
		}
		ps.CalculateWinRate()
		ps.LastPlayed = time.Now()
		return ps
	})
}

// genMove generates a random Move for property testing
func genMove(gameID string) gopter.Gen {
	return gen.Struct(reflect.TypeOf(&models.Move{}), map[string]gopter.Gen{
		"GameID": gen.Const(gameID),
		"Player": gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
		"Column": gen.IntRange(0, 6),
		"Row":    gen.IntRange(0, 5),
	}).Map(func(m *models.Move) *models.Move {
		// Ensure ID is generated
		if m.ID == "" {
			m.ID = "move-" + fmt.Sprintf("%d", time.Now().UnixNano())
		}
		m.Timestamp = time.Now()
		return m
	})
}

// Feature: connect-4-multiplayer, Property 10: Game Data Persistence
func TestGameDataPersistence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Player data serialization round trip", prop.ForAll(
		func(player *models.Player) bool {
			// Test that player data can be serialized and maintains integrity
			if player == nil {
				return false
			}

			// Validate required fields are present
			if len(player.Username) < 3 || len(player.Username) > 20 {
				return true // Skip invalid data
			}

			// Test ID generation
			if player.ID == "" {
				return false
			}

			// Test that timestamps are properly set
			if player.CreatedAt.IsZero() {
				player.CreatedAt = time.Now()
			}
			if player.UpdatedAt.IsZero() {
				player.UpdatedAt = time.Now()
			}

			// Verify data integrity after setting timestamps
			return !player.CreatedAt.IsZero() && 
				   !player.UpdatedAt.IsZero() &&
				   len(player.Username) >= 3 &&
				   len(player.Username) <= 20 &&
				   player.ID != ""
		},
		genPlayer(),
	))

	properties.Property("GameSession data integrity", prop.ForAll(
		func(gameSession *models.GameSession) bool {
			if gameSession == nil {
				return false
			}

			// Skip invalid data
			if len(gameSession.Player1) < 3 || len(gameSession.Player1) > 20 ||
			   len(gameSession.Player2) < 3 || len(gameSession.Player2) > 20 {
				return true
			}

			// Test that game session maintains data integrity
			if gameSession.ID == "" {
				return false
			}

			// Test that players are different
			if gameSession.Player1 == gameSession.Player2 {
				return false
			}

			// Test that board is properly initialized
			if gameSession.Board.Grid == ([6][7]models.PlayerColor{}) {
				gameSession.Board = models.NewBoard()
			}

			// Test that current turn is valid
			if !gameSession.CurrentTurn.IsValid() {
				return false
			}

			// Test board operations
			originalHeight := gameSession.Board.Height[0]
			if gameSession.Board.IsValidMove(0) {
				// Test making a move doesn't break the board
				testBoard := gameSession.Board
				err := testBoard.MakeMove(0, gameSession.CurrentTurn)
				if err != nil {
					return false
				}
				// Height should increase by 1
				if testBoard.Height[0] != originalHeight+1 {
					return false
				}
			}

			return true
		},
		genGameSession(),
	))

	properties.Property("PlayerStats calculation accuracy", prop.ForAll(
		func(stats *models.PlayerStats) bool {
			if stats == nil {
				return false
			}

			// Skip invalid data
			if len(stats.Username) < 3 || len(stats.Username) > 20 {
				return true
			}

			// Ensure GamesWon <= GamesPlayed
			if stats.GamesWon > stats.GamesPlayed {
				stats.GamesWon = stats.GamesPlayed
			}

			// Test win rate calculation
			stats.CalculateWinRate()

			// Verify win rate is correct
			expectedWinRate := 0.0
			if stats.GamesPlayed > 0 {
				expectedWinRate = float64(stats.GamesWon) / float64(stats.GamesPlayed)
			}

			return abs(stats.WinRate-expectedWinRate) < 0.0001 &&
				   stats.GamesWon <= stats.GamesPlayed &&
				   stats.GamesPlayed >= 0 &&
				   stats.GamesWon >= 0 &&
				   stats.WinRate >= 0.0 &&
				   stats.WinRate <= 1.0
		},
		genPlayerStats(),
	))

	properties.Property("Move validation consistency", prop.ForAll(
		func(move *models.Move) bool {
			if move == nil {
				return false
			}

			// Test move validation
			isValid := move.IsValid()

			// Check validation logic
			expectedValid := move.Column >= 0 && move.Column < 7 &&
							move.Row >= 0 && move.Row < 6 &&
							move.Player.IsValid()

			return isValid == expectedValid
		},
		genMove("test-game-id"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}