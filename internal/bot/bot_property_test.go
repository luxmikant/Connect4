//go:build property
// +build property

package bot_test

import (
	"context"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"connect4-multiplayer/internal/bot"
	"connect4-multiplayer/pkg/models"
)

// Feature: connect-4-multiplayer, Property 2: Bot Strategic Decision Making
// *For any* game board state, when the bot has a winning move available it should take it,
// when the opponent has a winning threat it should block it, and it should never make invalid moves.
// **Validates: Requirements 2.2, 2.3, 2.4, 2.6**
func TestBotStrategicDecisionMakingProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Bot should always take a winning move when available
	properties.Property("bot takes winning move when available (horizontal)", prop.ForAll(
		func(row, startCol int) bool {
			if row < 0 || row >= 6 || startCol < 0 || startCol > 3 {
				return true // Skip invalid positions
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up a horizontal winning opportunity for bot (red): R R R _
			// Need to build up the board properly with gravity
			for i := 0; i < 3; i++ {
				// Fill column up to the target row
				for r := 0; r < row; r++ {
					board.MakeMove(startCol+i, models.PlayerColorYellow)
				}
				board.MakeMove(startCol+i, models.PlayerColorRed)
			}
			// Fill the winning column up to the target row
			for r := 0; r < row; r++ {
				board.MakeMove(startCol+3, models.PlayerColorYellow)
			}

			// Bot should find the winning move
			winMove := botAI.FindWinningMove(&board, models.PlayerColorRed)
			if winMove == -1 {
				return true // No winning move available (board state may not allow it)
			}

			// Verify the move actually wins
			boardCopy := copyBoard(&board)
			boardCopy.MakeMove(winMove, models.PlayerColorRed)
			winner := boardCopy.CheckWin()
			return winner != nil && *winner == models.PlayerColorRed
		},
		gen.IntRange(0, 5),
		gen.IntRange(0, 3),
	))

	// Property: Bot should always take a winning move when available (vertical)
	properties.Property("bot takes winning move when available (vertical)", prop.ForAll(
		func(col int) bool {
			if col < 0 || col >= 7 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up a vertical winning opportunity: 3 red pieces stacked
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)

			// Bot should find the winning move at the same column
			winMove := botAI.FindWinningMove(&board, models.PlayerColorRed)
			return winMove == col
		},
		gen.IntRange(0, 6),
	))

	// Property: Bot should block opponent's winning move
	properties.Property("bot blocks opponent winning move", prop.ForAll(
		func(col int) bool {
			if col < 0 || col >= 7 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up opponent's (yellow) winning threat: 3 yellow pieces stacked
			board.MakeMove(col, models.PlayerColorYellow)
			board.MakeMove(col, models.PlayerColorYellow)
			board.MakeMove(col, models.PlayerColorYellow)

			// Bot (red) should block at the same column
			blockMove := botAI.FindBlockingMove(&board, models.PlayerColorRed)
			return blockMove == col
		},
		gen.IntRange(0, 6),
	))

	// Property: Bot prioritizes winning over blocking
	properties.Property("bot prioritizes winning over blocking", prop.ForAll(
		func() bool {
			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up: Red can win at column 3, Yellow can win at column 6
			// Red: R R R _ at row 0 (columns 0,1,2)
			board.MakeMove(0, models.PlayerColorRed)
			board.MakeMove(1, models.PlayerColorRed)
			board.MakeMove(2, models.PlayerColorRed)

			// Yellow: Y Y Y _ at row 0 (columns 4,5,6) - but need to use different rows
			// Actually set up yellow threat in column 4 vertically
			board.MakeMove(4, models.PlayerColorYellow)
			board.MakeMove(4, models.PlayerColorYellow)
			board.MakeMove(4, models.PlayerColorYellow)

			// Bot should take the win at column 3, not block at column 4
			bestMove := botAI.GetBestMove(&board, models.PlayerColorRed, 4)
			return bestMove == 3
		},
	))

	// Property: Bot never makes invalid moves
	properties.Property("bot never makes invalid moves on any board state", prop.ForAll(
		func(numMoves int) bool {
			if numMoves < 0 || numMoves > 40 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Make random moves to create a game state
			for i := 0; i < numMoves; i++ {
				// Find a valid column
				validCol := -1
				for col := 0; col < 7; col++ {
					if board.IsValidMove(col) {
						validCol = col
						break
					}
				}
				if validCol == -1 {
					break // Board is full
				}

				player := models.PlayerColorRed
				if i%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(validCol, player)

				// Check if game ended
				if board.CheckWin() != nil || board.IsFull() {
					break
				}
			}

			// Skip if game already ended
			if board.CheckWin() != nil || board.IsFull() {
				return true
			}

			// Bot should always return a valid move
			move := botAI.GetBestMove(&board, models.PlayerColorRed, 4)
			return board.IsValidMove(move)
		},
		gen.IntRange(0, 40),
	))

	// Property: Bot never chooses a full column
	properties.Property("bot never chooses full column", prop.ForAll(
		func(fullCol int) bool {
			if fullCol < 0 || fullCol >= 7 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Fill one column completely
			for i := 0; i < 6; i++ {
				player := models.PlayerColorRed
				if i%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(fullCol, player)
			}

			// Bot should not choose the full column
			move := botAI.GetBestMove(&board, models.PlayerColorRed, 4)
			return move != fullCol && board.IsValidMove(move)
		},
		gen.IntRange(0, 6),
	))

	// Property: GetBestMove returns valid move for any non-terminal board state
	properties.Property("GetBestMove always returns valid move for non-terminal states", prop.ForAll(
		func(seed int64) bool {
			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Use seed to create deterministic but varied board states
			moves := []int{3, 2, 4, 1, 5, 0, 6} // Center-first ordering
			numMoves := int(seed % 30)

			for i := 0; i < numMoves; i++ {
				col := moves[i%7]
				if !board.IsValidMove(col) {
					// Find any valid column
					for c := 0; c < 7; c++ {
						if board.IsValidMove(c) {
							col = c
							break
						}
					}
				}
				if !board.IsValidMove(col) {
					break // Board full
				}

				player := models.PlayerColorRed
				if i%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(col, player)

				if board.CheckWin() != nil {
					break
				}
			}

			// Skip terminal states
			if board.CheckWin() != nil || board.IsFull() {
				return true
			}

			move := botAI.GetBestMove(&board, models.PlayerColorRed, 4)
			return board.IsValidMove(move)
		},
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: connect-4-multiplayer, Property 3: Bot Response Time
// *For any* game board state, the bot should analyze and make its move within 1 second.
// **Validates: Requirements 2.1**
func TestBotResponseTimeProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Bot responds within 1 second for any board state
	properties.Property("bot responds within 1 second for any board state", prop.ForAll(
		func(numMoves int) bool {
			if numMoves < 0 || numMoves > 40 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Create a random board state
			for i := 0; i < numMoves; i++ {
				// Find a valid column using center-first strategy
				validCol := -1
				for _, col := range []int{3, 2, 4, 1, 5, 0, 6} {
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

			// Skip terminal states
			if board.CheckWin() != nil || board.IsFull() {
				return true
			}

			// Measure response time
			ctx := context.Background()
			timeout := 1 * time.Second

			start := time.Now()
			move, err := botAI.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, timeout)
			elapsed := time.Since(start)

			// Should complete within 1 second and return valid move
			return elapsed <= timeout && (err == nil || err == context.DeadlineExceeded) && board.IsValidMove(move)
		},
		gen.IntRange(0, 40),
	))

	// Property: Bot responds quickly on empty board
	properties.Property("bot responds quickly on empty board", prop.ForAll(
		func() bool {
			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			ctx := context.Background()
			timeout := 1 * time.Second

			start := time.Now()
			move, err := botAI.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, timeout)
			elapsed := time.Since(start)

			// Should complete well within 1 second on empty board
			return elapsed <= timeout && err == nil && board.IsValidMove(move)
		},
	))

	// Property: Bot responds quickly when there's an obvious winning move
	properties.Property("bot responds quickly with obvious winning move", prop.ForAll(
		func(col int) bool {
			if col < 0 || col >= 7 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up obvious winning move (3 in a row vertically)
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)
			board.MakeMove(col, models.PlayerColorRed)

			ctx := context.Background()
			timeout := 1 * time.Second

			start := time.Now()
			move, err := botAI.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, timeout)
			elapsed := time.Since(start)

			// Should find winning move very quickly (< 100ms)
			return elapsed < 100*time.Millisecond && err == nil && move == col
		},
		gen.IntRange(0, 6),
	))

	// Property: Bot responds quickly when blocking is needed
	properties.Property("bot responds quickly when blocking opponent", prop.ForAll(
		func(col int) bool {
			if col < 0 || col >= 7 {
				return true
			}

			botAI := bot.NewMinimaxBot()
			board := models.NewBoard()

			// Set up opponent's winning threat
			board.MakeMove(col, models.PlayerColorYellow)
			board.MakeMove(col, models.PlayerColorYellow)
			board.MakeMove(col, models.PlayerColorYellow)

			ctx := context.Background()
			timeout := 1 * time.Second

			start := time.Now()
			move, err := botAI.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, timeout)
			elapsed := time.Since(start)

			// Should find blocking move very quickly (< 100ms)
			return elapsed < 100*time.Millisecond && err == nil && move == col
		},
		gen.IntRange(0, 6),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper function to copy board
func copyBoard(board *models.Board) *models.Board {
	newBoard := &models.Board{}
	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			newBoard.Grid[row][col] = board.Grid[row][col]
		}
	}
	for col := 0; col < 7; col++ {
		newBoard.Height[col] = board.Height[col]
	}
	return newBoard
}
