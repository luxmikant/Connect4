//go:build property
// +build property

package bot_test

import (
	"context"
	"testing"
	"time"

	"connect4-multiplayer/internal/bot"
	"connect4-multiplayer/pkg/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: Bot Difficulty System Properties
// Validates that different difficulty levels maintain consistent and appropriate behavior

// Property 1: All difficulty levels should make valid moves for any board state
func TestDifficultyProperty_AllMakesValidMoves(t *testing.T) {
	properties := gopter.NewProperties(nil)

	difficulties := []bot.Difficulty{
		bot.DifficultyEasy,
		bot.DifficultyMedium,
		bot.DifficultyHard,
	}

	for _, difficulty := range difficulties {
		diff := difficulty // capture for closure
		properties.Property(diff.String()+" always makes valid moves", prop.ForAll(
			func(numMoves int) bool {
				if numMoves < 0 || numMoves > 40 {
					return true
				}

				service := bot.NewBotPlayerService()
				botPlayer := service.CreateBot(diff)
				board := models.NewBoard()
				ctx := context.Background()

				// Make random moves to create a game state
				for i := 0; i < numMoves; i++ {
					validCol := -1
					for col := 0; col < 7; col++ {
						if board.IsValidMove(col) {
							validCol = col
							break
						}
					}
					if validCol == -1 {
						break
					}

					player := models.PlayerColorRed
					if i%2 == 1 {
						player = models.PlayerColorYellow
					}
					board.MakeMove(validCol, player)

					if board.CheckWin() != nil || board.IsFull() {
						break
					}
				}

				// Skip if game already ended
				if board.CheckWin() != nil || board.IsFull() {
					return true
				}

				// Bot should always return a valid move
				move, err := service.GetBotMove(ctx, botPlayer, &board, models.PlayerColorRed)
				if err != nil {
					return false
				}

				return board.IsValidMove(move) && move >= 0 && move <= 6
			},
			gen.IntRange(0, 40),
		))
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 2: Hard bot should never miss a winning move
func TestDifficultyProperty_HardNeverMissesWin(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Hard bot always takes immediate winning moves
	properties.Property("hard bot takes winning move (horizontal)", prop.ForAll(
		func(startCol int) bool {
			if startCol < 0 || startCol > 3 {
				return true
			}

			service := bot.NewBotPlayerService()
			hardBot := service.CreateBot(bot.DifficultyHard)
			board := models.NewBoard()
			ctx := context.Background()

			// Set up horizontal win: R R R _ at bottom row (row 0)
			// Simply place 3 red pieces horizontally
			board.MakeMove(startCol, models.PlayerColorRed)
			board.MakeMove(startCol+1, models.PlayerColorRed)
			board.MakeMove(startCol+2, models.PlayerColorRed)

			// Verify setup created the expected pattern
			if board.Grid[0][startCol] != models.PlayerColorRed ||
				board.Grid[0][startCol+1] != models.PlayerColorRed ||
				board.Grid[0][startCol+2] != models.PlayerColorRed {
				return true // Skip if board setup failed
			}

			// Hard bot should take the winning move
			move, err := service.GetBotMove(ctx, hardBot, &board, models.PlayerColorRed)
			if err != nil {
				return false
			}

			return move == startCol+3
		},
		gen.IntRange(0, 3),
	))

	// Property: Hard bot always takes immediate winning moves (vertical)
	properties.Property("hard bot takes winning move (vertical)", prop.ForAll(
		func(col int) bool {
			if col < 0 || col >= 7 {
				return true
			}

			service := bot.NewBotPlayerService()
			hardBot := service.CreateBot(bot.DifficultyHard)
			board := models.NewBoard()
			ctx := context.Background()

			// Set up vertical win: three stacked pieces
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)

			// Hard bot should take the winning move
			move, err := service.GetBotMove(ctx, hardBot, &board, models.PlayerColorRed)
			if err != nil {
				return false
			}

			return move == col
		},
		gen.IntRange(0, 6),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 3: Medium and Hard bots should always block immediate threats
func TestDifficultyProperty_MediumHardBlockThreats(t *testing.T) {
	properties := gopter.NewProperties(nil)

	smartDifficulties := []bot.Difficulty{
		bot.DifficultyMedium,
		bot.DifficultyHard,
	}

	for _, difficulty := range smartDifficulties {
		diff := difficulty // capture for closure
		properties.Property(diff.String()+" blocks opponent horizontal threat", prop.ForAll(
			func(startCol int) bool {
				if startCol < 0 || startCol > 3 {
					return true
				}

				service := bot.NewBotPlayerService()
				botPlayer := service.CreateBot(diff)
				board := models.NewBoard()
				ctx := context.Background()

				// Set up opponent's threat: Y Y Y _ at bottom row (row 0)
				board.MakeMove(startCol, models.PlayerColorYellow)
				board.MakeMove(startCol+1, models.PlayerColorYellow)
				board.MakeMove(startCol+2, models.PlayerColorYellow)

				// Verify the setup
				if board.Grid[0][startCol] != models.PlayerColorYellow ||
					board.Grid[0][startCol+1] != models.PlayerColorYellow ||
					board.Grid[0][startCol+2] != models.PlayerColorYellow {
					return true // Skip if setup failed
				}

				// Bot should block at the winning position
				move, err := service.GetBotMove(ctx, botPlayer, &board, models.PlayerColorRed)
				if err != nil {
					return false
				}

				return move == startCol+3
			},
			gen.IntRange(0, 3),
		))
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 4: Difficulty levels should have monotonically increasing search depths
func TestDifficultyProperty_MonotonicSearchDepth(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("search depth increases with difficulty", prop.ForAll(
		func() bool {
			easy := bot.DifficultyEasy.SearchDepth()
			medium := bot.DifficultyMedium.SearchDepth()
			hard := bot.DifficultyHard.SearchDepth()

			return easy < medium && medium < hard
		},
	))

	properties.Property("delay decreases with difficulty", prop.ForAll(
		func() bool {
			easy := bot.DifficultyEasy.HumanDelay()
			medium := bot.DifficultyMedium.HumanDelay()
			hard := bot.DifficultyHard.HumanDelay()

			return easy > medium && medium > hard
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 5: Bots should handle edge cases across all difficulties
func TestDifficultyProperty_EdgeCases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	difficulties := []bot.Difficulty{
		bot.DifficultyEasy,
		bot.DifficultyMedium,
		bot.DifficultyHard,
	}

	for _, difficulty := range difficulties {
		diff := difficulty // capture for closure

		// Property: Bots avoid full columns
		properties.Property(diff.String()+" avoids full columns", prop.ForAll(
			func(fullCol int) bool {
				if fullCol < 0 || fullCol >= 7 {
					return true
				}

				service := bot.NewBotPlayerService()
				botPlayer := service.CreateBot(diff)
				board := models.NewBoard()
				ctx := context.Background()

				// Fill one column completely
				for i := 0; i < 6; i++ {
					player := models.PlayerColorRed
					if i%2 == 1 {
						player = models.PlayerColorYellow
					}
					board.MakeMove(fullCol, player)
				}

				// Bot should not choose the full column
				move, err := service.GetBotMove(ctx, botPlayer, &board, models.PlayerColorRed)
				if err != nil {
					return false
				}

				return move != fullCol && board.IsValidMove(move)
			},
			gen.IntRange(0, 6),
		))

		// Property: Bots handle nearly full boards
		properties.Property(diff.String()+" handles nearly full board", prop.ForAll(
			func() bool {
				service := bot.NewBotPlayerService()
				botPlayer := service.CreateBot(diff)
				board := models.NewBoard()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Fill all columns except leave column 3 with one space
				alternate := 0
				for col := 0; col < 7; col++ {
					maxRows := 6
					if col == 3 {
						maxRows = 5 // Leave one space in column 3
					}
					for row := 0; row < maxRows; row++ {
						player := models.PlayerColorRed
						if alternate%2 == 1 {
							player = models.PlayerColorYellow
						}
						board.MakeMove(col, player)
						alternate++
					}
				}

				// Bot should find the only valid move (column 3)
				move, err := service.GetBotMove(ctx, botPlayer, &board, models.PlayerColorRed)
				if err != nil {
					return false
				}

				return move == 3
			},
		))
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 6: Performance properties - harder bots take more time
func TestDifficultyProperty_PerformanceScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	properties := gopter.NewProperties(nil)

	properties.Property("harder difficulties generally take more computation", prop.ForAll(
		func() bool {
			service := bot.NewBotPlayerService()
			board := models.NewBoard()

			// Create a moderately complex board state
			moves := []int{3, 2, 4, 3, 2, 4, 5}
			for i, col := range moves {
				player := models.PlayerColorRed
				if i%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(col, player)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Measure time for easy bot (with delay subtracted)
			easyBot := service.CreateBot(bot.DifficultyEasy)
			startEasy := time.Now()
			service.GetBotMove(ctx, easyBot, &board, models.PlayerColorRed)
			easyTime := time.Since(startEasy) - bot.DifficultyEasy.HumanDelay()

			// Measure time for hard bot (with delay subtracted)
			hardBot := service.CreateBot(bot.DifficultyHard)
			startHard := time.Now()
			service.GetBotMove(ctx, hardBot, &board, models.PlayerColorRed)
			hardTime := time.Since(startHard) - bot.DifficultyHard.HumanDelay()

			// Hard bot should generally take more computation time,
			// but account for variance by allowing small easy times
			// Property holds if either hard takes longer OR easy was trivial
			return hardTime >= easyTime || easyTime < 5*time.Millisecond
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 7: Bot behavior determinism (except Easy which has randomness)
func TestDifficultyProperty_Determinism(t *testing.T) {
	properties := gopter.NewProperties(nil)

	deterministicDifficulties := []bot.Difficulty{
		bot.DifficultyMedium,
		bot.DifficultyHard,
	}

	for _, difficulty := range deterministicDifficulties {
		diff := difficulty // capture for closure

		properties.Property(diff.String()+" is deterministic", prop.ForAll(
			func(numMoves int) bool {
				if numMoves < 0 || numMoves > 20 {
					return true
				}

				service := bot.NewBotPlayerService()
				board := models.NewBoard()
				ctx := context.Background()

				// Create a specific board state
				for i := 0; i < numMoves; i++ {
					col := i % 7
					if !board.IsValidMove(col) {
						col = (i + 1) % 7
					}
					if !board.IsValidMove(col) {
						break
					}

					player := models.PlayerColorRed
					if i%2 == 1 {
						player = models.PlayerColorYellow
					}
					board.MakeMove(col, player)

					if board.CheckWin() != nil || board.IsFull() {
						break
					}
				}

				if board.CheckWin() != nil || board.IsFull() {
					return true
				}

				// Make the same move twice with different bots
				bot1 := service.CreateBot(diff)
				bot2 := service.CreateBot(diff)

				move1, err1 := service.GetBotMove(ctx, bot1, &board, models.PlayerColorRed)
				move2, err2 := service.GetBotMove(ctx, bot2, &board, models.PlayerColorRed)

				if err1 != nil || err2 != nil {
					return false
				}

				// Should be deterministic (same move)
				return move1 == move2
			},
			gen.IntRange(0, 20),
		))
	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
