package game_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/pkg/models"
)

// Helper function to create a test engine
func createTestEngine() (game.Engine, *MockGameSessionRepository, *MockMoveRepository) {
	gameRepo := NewMockGameSessionRepository()
	moveRepo := NewMockMoveRepository()
	engine := game.NewEngine(gameRepo, moveRepo)
	return engine, gameRepo, moveRepo
}

// =============================================================================
// Game Creation Tests
// =============================================================================

func TestCreateGame_Success(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, err := engine.CreateGame(ctx, "player1", "player2")

	require.NoError(t, err)
	assert.NotEmpty(t, gameSession.ID)
	assert.Equal(t, "player1", gameSession.Player1)
	assert.Equal(t, "player2", gameSession.Player2)
	assert.Equal(t, models.StatusInProgress, gameSession.Status)
	assert.Equal(t, models.PlayerColorRed, gameSession.CurrentTurn)
}

func TestCreateGame_EmptyPlayer1(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	_, err := engine.CreateGame(ctx, "", "player2")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestCreateGame_EmptyPlayer2(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	_, err := engine.CreateGame(ctx, "player1", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestCreateGame_SamePlayer(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	_, err := engine.CreateGame(ctx, "player1", "player1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "different usernames")
}

// =============================================================================
// Move Validation Tests - Requirements 5.1, 5.2
// =============================================================================

func TestValidateMove_ValidColumn(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Test all valid columns (0-6)
	for col := 0; col < 7; col++ {
		err := engine.ValidateMove(ctx, gameSession.ID, "player1", col)
		assert.NoError(t, err, "Column %d should be valid", col)
	}
}

func TestValidateMove_NegativeColumn(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	err := engine.ValidateMove(ctx, gameSession.ID, "player1", -1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column")
}

func TestValidateMove_ColumnTooHigh(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	err := engine.ValidateMove(ctx, gameSession.ID, "player1", 7)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid column")
}

func TestValidateMove_FullColumn(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Fill column 3 completely
	for row := 0; row < 6; row++ {
		gameSession.Board.Grid[row][3] = models.PlayerColorRed
	}
	gameSession.Board.Height[3] = 6
	gameRepo.Update(ctx, gameSession)

	err := engine.ValidateMove(ctx, gameSession.ID, "player1", 3)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "column 3 is full")
}

func TestValidateMove_WrongPlayerTurn(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
	// Player1 (red) starts first

	err := engine.ValidateMove(ctx, gameSession.ID, "player2", 3)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not player2's turn")
}

func TestValidateMove_GameNotActive(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
	gameSession.Status = models.StatusCompleted
	gameRepo.Update(ctx, gameSession)

	err := engine.ValidateMove(ctx, gameSession.ID, "player1", 3)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

// =============================================================================
// Make Move Tests - Requirements 5.1, 5.2
// =============================================================================

func TestMakeMove_PlacesDiscInLowestPosition(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	result, err := engine.MakeMove(ctx, gameSession.ID, "player1", 3)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Move.Row) // First disc lands at row 0
	assert.Equal(t, 3, result.Move.Column)
	assert.Equal(t, models.PlayerColorRed, result.GameSession.Board.Grid[0][3])
}

func TestMakeMove_StacksDiscsCorrectly(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Player1 moves in column 3
	result1, _ := engine.MakeMove(ctx, gameSession.ID, "player1", 3)
	assert.Equal(t, 0, result1.Move.Row)

	// Player2 moves in same column 3
	result2, _ := engine.MakeMove(ctx, gameSession.ID, "player2", 3)
	assert.Equal(t, 1, result2.Move.Row)

	// Player1 moves in same column 3 again
	result3, _ := engine.MakeMove(ctx, gameSession.ID, "player1", 3)
	assert.Equal(t, 2, result3.Move.Row)
}

func TestMakeMove_SwitchesTurns(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
	assert.Equal(t, models.PlayerColorRed, gameSession.CurrentTurn)

	result1, _ := engine.MakeMove(ctx, gameSession.ID, "player1", 3)
	assert.Equal(t, models.PlayerColorYellow, result1.GameSession.CurrentTurn)

	result2, _ := engine.MakeMove(ctx, gameSession.ID, "player2", 4)
	assert.Equal(t, models.PlayerColorRed, result2.GameSession.CurrentTurn)
}

func TestMakeMove_InvalidColumn(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	_, err := engine.MakeMove(ctx, gameSession.ID, "player1", -1)

	assert.Error(t, err)
}

// =============================================================================
// Win Detection Tests - Requirements 5.3
// =============================================================================

func TestCheckGameEnd_HorizontalWin_BottomRow(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create horizontal win at bottom row (row 0), columns 0-3
	for col := 0; col < 4; col++ {
		gameSession.Board.Grid[0][col] = models.PlayerColorRed
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorRed, *result.Winner)
	assert.False(t, result.IsDraw)
}

func TestCheckGameEnd_HorizontalWin_TopRow(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create horizontal win at top row (row 5), columns 3-6
	for col := 3; col < 7; col++ {
		gameSession.Board.Grid[5][col] = models.PlayerColorYellow
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorYellow, *result.Winner)
}

func TestCheckGameEnd_VerticalWin_LeftColumn(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create vertical win in column 0, rows 0-3
	for row := 0; row < 4; row++ {
		gameSession.Board.Grid[row][0] = models.PlayerColorRed
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorRed, *result.Winner)
}

func TestCheckGameEnd_VerticalWin_RightColumn(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create vertical win in column 6, rows 2-5
	for row := 2; row < 6; row++ {
		gameSession.Board.Grid[row][6] = models.PlayerColorYellow
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorYellow, *result.Winner)
}

func TestCheckGameEnd_DiagonalWin_BottomLeftToTopRight(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create diagonal win from (0,0) to (3,3)
	for i := 0; i < 4; i++ {
		gameSession.Board.Grid[i][i] = models.PlayerColorRed
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorRed, *result.Winner)
}

func TestCheckGameEnd_DiagonalWin_TopRightToBottomLeft(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create diagonal win from (0,6) to (3,3)
	for i := 0; i < 4; i++ {
		gameSession.Board.Grid[i][6-i] = models.PlayerColorYellow
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorYellow, *result.Winner)
}

func TestCheckGameEnd_DiagonalWin_UpperRight(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create diagonal win from (2,3) to (5,6)
	for i := 0; i < 4; i++ {
		gameSession.Board.Grid[2+i][3+i] = models.PlayerColorRed
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorRed, *result.Winner)
}

// =============================================================================
// Draw Detection Tests - Requirements 5.4
// =============================================================================

func TestCheckGameEnd_Draw_FullBoardNoWinner(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create a full board with no winner using a pattern that breaks all 4-in-a-row
	pattern := [6][7]models.PlayerColor{
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
	}
	gameSession.Board.Grid = pattern
	for col := 0; col < 7; col++ {
		gameSession.Board.Height[col] = 6
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.Nil(t, result.Winner)
	assert.True(t, result.IsDraw)
	assert.Equal(t, "board_full", result.Reason)
}

func TestCheckGameEnd_NoWinNoDrawGameContinues(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Make a few moves but no win or draw
	gameSession.Board.Grid[0][0] = models.PlayerColorRed
	gameSession.Board.Grid[0][1] = models.PlayerColorYellow
	gameSession.Board.Grid[0][2] = models.PlayerColorRed
	gameSession.Board.Height[0] = 1
	gameSession.Board.Height[1] = 1
	gameSession.Board.Height[2] = 1

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.False(t, result.GameEnded)
	assert.Nil(t, result.Winner)
	assert.False(t, result.IsDraw)
	assert.Equal(t, "game_in_progress", result.Reason)
}

func TestCheckGameEnd_ThreeInARowNotWin(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create 3 in a row (not 4)
	for col := 0; col < 3; col++ {
		gameSession.Board.Grid[0][col] = models.PlayerColorRed
	}
	gameRepo.Update(ctx, gameSession)

	result, err := engine.CheckGameEnd(ctx, gameSession)

	require.NoError(t, err)
	assert.False(t, result.GameEnded)
	assert.Nil(t, result.Winner)
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestMakeMove_WinEndsGame(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Set up 3 in a row for player1
	gameSession.Board.Grid[0][0] = models.PlayerColorRed
	gameSession.Board.Grid[0][1] = models.PlayerColorRed
	gameSession.Board.Grid[0][2] = models.PlayerColorRed
	gameSession.Board.Height[0] = 1
	gameSession.Board.Height[1] = 1
	gameSession.Board.Height[2] = 1
	gameRepo.Update(ctx, gameSession)

	// Player1 makes winning move
	result, err := engine.MakeMove(ctx, gameSession.ID, "player1", 3)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.NotNil(t, result.Winner)
	assert.Equal(t, models.PlayerColorRed, *result.Winner)
	assert.Equal(t, models.StatusCompleted, result.GameSession.Status)
}

func TestMakeMove_DrawEndsGame(t *testing.T) {
	engine, gameRepo, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	// Create almost full board with no winner (one spot left)
	pattern := [6][7]models.PlayerColor{
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
		{models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
		{models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, ""},
	}
	gameSession.Board.Grid = pattern
	for col := 0; col < 6; col++ {
		gameSession.Board.Height[col] = 6
	}
	gameSession.Board.Height[6] = 5 // Column 6 has one spot left
	gameRepo.Update(ctx, gameSession)

	// Player1 makes the final move
	result, err := engine.MakeMove(ctx, gameSession.ID, "player1", 6)

	require.NoError(t, err)
	assert.True(t, result.GameEnded)
	assert.True(t, result.IsDraw)
	assert.Nil(t, result.Winner)
	assert.Equal(t, models.StatusCompleted, result.GameSession.Status)
}

func TestIsPlayerTurn_Player1First(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

	assert.True(t, engine.IsPlayerTurn(ctx, gameSession, "player1"))
	assert.False(t, engine.IsPlayerTurn(ctx, gameSession, "player2"))
}

func TestGetGame_NotFound(t *testing.T) {
	engine, _, _ := createTestEngine()
	ctx := context.Background()

	_, err := engine.GetGame(ctx, "non-existent-id")

	assert.Error(t, err)
}

// =============================================================================
// Boundary Position Win Tests
// =============================================================================

func TestCheckGameEnd_HorizontalWin_AllPositions(t *testing.T) {
	testCases := []struct {
		name     string
		row      int
		startCol int
	}{
		{"bottom-left", 0, 0},
		{"bottom-middle", 0, 1},
		{"bottom-right", 0, 3},
		{"top-left", 5, 0},
		{"top-right", 5, 3},
		{"middle", 3, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, gameRepo, _ := createTestEngine()
			ctx := context.Background()

			gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

			for col := tc.startCol; col < tc.startCol+4; col++ {
				gameSession.Board.Grid[tc.row][col] = models.PlayerColorRed
			}
			gameRepo.Update(ctx, gameSession)

			result, err := engine.CheckGameEnd(ctx, gameSession)

			require.NoError(t, err)
			assert.True(t, result.GameEnded, "Expected win at row %d, cols %d-%d", tc.row, tc.startCol, tc.startCol+3)
			assert.NotNil(t, result.Winner)
			assert.Equal(t, models.PlayerColorRed, *result.Winner)
		})
	}
}

func TestCheckGameEnd_VerticalWin_AllPositions(t *testing.T) {
	testCases := []struct {
		name     string
		startRow int
		col      int
	}{
		{"left-bottom", 0, 0},
		{"left-top", 2, 0},
		{"right-bottom", 0, 6},
		{"right-top", 2, 6},
		{"center-bottom", 0, 3},
		{"center-top", 2, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, gameRepo, _ := createTestEngine()
			ctx := context.Background()

			gameSession, _ := engine.CreateGame(ctx, "player1", "player2")

			for row := tc.startRow; row < tc.startRow+4; row++ {
				gameSession.Board.Grid[row][tc.col] = models.PlayerColorYellow
			}
			gameRepo.Update(ctx, gameSession)

			result, err := engine.CheckGameEnd(ctx, gameSession)

			require.NoError(t, err)
			assert.True(t, result.GameEnded, "Expected win at col %d, rows %d-%d", tc.col, tc.startRow, tc.startRow+3)
			assert.NotNil(t, result.Winner)
			assert.Equal(t, models.PlayerColorYellow, *result.Winner)
		})
	}
}
