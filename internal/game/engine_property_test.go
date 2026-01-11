//go:build property
// +build property

package game_test

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/pkg/models"
)

// Feature: connect-4-multiplayer, Property 7: Game Move Validation and Physics
func TestMoveValidationProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid moves should be accepted for non-full columns", prop.ForAll(
		func(column int) bool {
			// Create a fresh game
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			// Valid column should be accepted if not full
			if gameSession.Board.IsValidMove(column) {
				err := engine.ValidateMove(ctx, gameSession.ID, "player1", column)
				return err == nil
			}
			return true // Skip if column is already full
		},
		gen.IntRange(0, 6),
	))

	properties.Property("negative columns should be rejected", prop.ForAll(
		func(column int) bool {
			// Create a fresh game
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			// Invalid column should be rejected
			err = engine.ValidateMove(ctx, gameSession.ID, "player1", column)
			return err != nil
		},
		gen.IntRange(-100, -1),
	))

	properties.Property("columns >= 7 should be rejected", prop.ForAll(
		func(column int) bool {
			// Create a fresh game
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			// Invalid column should be rejected
			err = engine.ValidateMove(ctx, gameSession.ID, "player1", column)
			return err != nil
		},
		gen.IntRange(7, 100),
	))

	properties.Property("moves should place discs in lowest available position", prop.ForAll(
		func(column int) bool {
			// Create a fresh game
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			if !gameSession.Board.IsValidMove(column) {
				return true // Skip invalid moves
			}
			
			// Record the expected row (lowest available)
			expectedRow := gameSession.Board.Height[column]
			
			// Make the move
			result, err := engine.MakeMove(ctx, gameSession.ID, "player1", column)
			if err != nil {
				return false
			}
			
			// Check that the disc was placed in the expected row
			return result.Move.Row == expectedRow && 
				   result.GameSession.Board.Grid[expectedRow][column] == models.PlayerColorRed
		},
		gen.IntRange(0, 6),
	))

	properties.Property("players can only move on their turn", prop.ForAll(
		func(column int) bool {
			// Create a fresh game
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			if !gameSession.Board.IsValidMove(column) {
				return true // Skip invalid moves
			}
			
			// Player2 should not be able to move when it's Player1's turn
			err = engine.ValidateMove(ctx, gameSession.ID, "player2", column)
			return err != nil
		},
		gen.IntRange(0, 6),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: connect-4-multiplayer, Property 8: Win and Draw Detection
func TestWinAndDrawDetectionProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Horizontal wins should be detected correctly
	properties.Property("horizontal 4-in-a-row should be detected as win", prop.ForAll(
		func(row, startCol int, player models.PlayerColor) bool {
			if row < 0 || row >= 6 || startCol < 0 || startCol > 3 {
				return true // Skip invalid positions
			}
			
			// Create empty board
			board := models.NewBoard()
			
			// Place 4 consecutive discs horizontally
			for i := 0; i < 4; i++ {
				board.Grid[row][startCol+i] = player
			}
			
			// Check that win is detected
			winner := board.CheckWin()
			return winner != nil && *winner == player
		},
		gen.IntRange(0, 5),    // row
		gen.IntRange(0, 3),    // startCol (0-3 to fit 4 consecutive)
		gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
	))

	// Property: Vertical wins should be detected correctly
	properties.Property("vertical 4-in-a-row should be detected as win", prop.ForAll(
		func(startRow, col int, player models.PlayerColor) bool {
			if startRow < 0 || startRow > 2 || col < 0 || col >= 7 {
				return true // Skip invalid positions
			}
			
			// Create empty board
			board := models.NewBoard()
			
			// Place 4 consecutive discs vertically
			for i := 0; i < 4; i++ {
				board.Grid[startRow+i][col] = player
			}
			
			// Check that win is detected
			winner := board.CheckWin()
			return winner != nil && *winner == player
		},
		gen.IntRange(0, 2),    // startRow (0-2 to fit 4 consecutive)
		gen.IntRange(0, 6),    // col
		gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
	))

	// Property: Diagonal wins (top-left to bottom-right) should be detected correctly
	properties.Property("diagonal 4-in-a-row (TL-BR) should be detected as win", prop.ForAll(
		func(startRow, startCol int, player models.PlayerColor) bool {
			if startRow < 0 || startRow > 2 || startCol < 0 || startCol > 3 {
				return true // Skip invalid positions
			}
			
			// Create empty board
			board := models.NewBoard()
			
			// Place 4 consecutive discs diagonally (top-left to bottom-right)
			for i := 0; i < 4; i++ {
				board.Grid[startRow+i][startCol+i] = player
			}
			
			// Check that win is detected
			winner := board.CheckWin()
			return winner != nil && *winner == player
		},
		gen.IntRange(0, 2),    // startRow (0-2 to fit 4 consecutive)
		gen.IntRange(0, 3),    // startCol (0-3 to fit 4 consecutive)
		gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
	))

	// Property: Diagonal wins (top-right to bottom-left) should be detected correctly
	properties.Property("diagonal 4-in-a-row (TR-BL) should be detected as win", prop.ForAll(
		func(startRow, startCol int, player models.PlayerColor) bool {
			if startRow < 0 || startRow > 2 || startCol < 3 || startCol >= 7 {
				return true // Skip invalid positions
			}
			
			// Create empty board
			board := models.NewBoard()
			
			// Place 4 consecutive discs diagonally (top-right to bottom-left)
			for i := 0; i < 4; i++ {
				board.Grid[startRow+i][startCol-i] = player
			}
			
			// Check that win is detected
			winner := board.CheckWin()
			return winner != nil && *winner == player
		},
		gen.IntRange(0, 2),    // startRow (0-2 to fit 4 consecutive)
		gen.IntRange(3, 6),    // startCol (3-6 to fit 4 consecutive going left)
		gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
	))

	// Property: Full boards without 4-in-a-row should be detected as draw
	properties.Property("full board without winner should be detected as draw", prop.ForAll(
		func() bool {
			// Create a full board with a carefully crafted pattern that prevents all 4-in-a-row
			board := models.NewBoard()
			
			// Use a pattern that breaks every possible 4-in-a-row
			// Pattern ensures max 3 consecutive in any direction
			pattern := [6][7]models.PlayerColor{
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
			}
			
			// Copy pattern to board
			board.Grid = pattern
			
			// Set all columns as full
			for col := 0; col < 7; col++ {
				board.Height[col] = 6
			}
			
			// Verify no winner exists and board is full
			winner := board.CheckWin()
			return winner == nil && board.IsFull()
		},
	))

	// Property: Empty or partially filled boards should not detect winner or draw
	properties.Property("non-full boards without 4-in-a-row should not detect winner or draw", prop.ForAll(
		func(numMoves int) bool {
			if numMoves < 0 || numMoves >= 42 {
				return true // Skip invalid move counts
			}
			
			// Create empty board
			board := models.NewBoard()
			
			// Make random moves without creating 4-in-a-row
			moveCount := 0
			for col := 0; col < 7 && moveCount < numMoves; col++ {
				for row := 0; row < 3 && moveCount < numMoves; row++ { // Only fill bottom 3 rows to avoid vertical wins
					player := models.PlayerColorRed
					if moveCount%2 == 1 {
						player = models.PlayerColorYellow
					}
					board.Grid[row][col] = player
					board.Height[col] = row + 1
					moveCount++
				}
			}
			
			// Verify no winner and not full
			winner := board.CheckWin()
			return winner == nil && !board.IsFull()
		},
		gen.IntRange(0, 20), // Limit moves to avoid filling board
	))

	// Property: Game engine correctly identifies game end conditions
	properties.Property("game engine correctly identifies win conditions", prop.ForAll(
		func(row, startCol int, player models.PlayerColor) bool {
			if row < 0 || row >= 6 || startCol < 0 || startCol > 3 {
				return true // Skip invalid positions
			}
			
			// Create game engine
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			// Create a winning board (horizontal win)
			for i := 0; i < 4; i++ {
				gameSession.Board.Grid[row][startCol+i] = player
			}
			
			// Check game end detection
			result, err := engine.CheckGameEnd(ctx, gameSession)
			if err != nil {
				return false
			}
			
			// Should detect win correctly
			return result.GameEnded && result.Winner != nil && *result.Winner == player && !result.IsDraw
		},
		gen.IntRange(0, 5),    // row
		gen.IntRange(0, 3),    // startCol
		gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
	))

	// Property: Game engine correctly identifies draw conditions
	properties.Property("game engine correctly identifies draw conditions", prop.ForAll(
		func() bool {
			// Create game engine
			gameRepo := NewMockGameSessionRepository()
			moveRepo := NewMockMoveRepository()
			engine := game.NewEngine(gameRepo, moveRepo)
			
			ctx := context.Background()
			gameSession, err := engine.CreateGame(ctx, "player1", "player2")
			if err != nil {
				return false
			}
			
			// Create a full board without winner using the same safe pattern
			pattern := [6][7]models.PlayerColor{
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
				{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
				{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
			}
			
			// Copy pattern to board
			gameSession.Board.Grid = pattern
			
			// Set all columns as full
			for col := 0; col < 7; col++ {
				gameSession.Board.Height[col] = 6
			}
			
			// Check game end detection
			result, err := engine.CheckGameEnd(ctx, gameSession)
			if err != nil {
				return false
			}
			
			// Should detect draw correctly
			return result.GameEnded && result.Winner == nil && result.IsDraw
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}