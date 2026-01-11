package bot

import (
	"context"
	"testing"
	"time"

	"connect4-multiplayer/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMinimaxBot(t *testing.T) {
	bot := NewMinimaxBot()
	assert.NotNil(t, bot)
}

func TestFindWinningMove_Horizontal(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Set up a horizontal winning opportunity for red: R R R _
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(1, models.PlayerColorRed)
	board.MakeMove(2, models.PlayerColorRed)

	// Bot should find the winning move at column 3
	winMove := bot.FindWinningMove(&board, models.PlayerColorRed)
	assert.Equal(t, 3, winMove)
}

func TestFindWinningMove_Vertical(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Set up a vertical winning opportunity for red in column 0
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(0, models.PlayerColorRed)

	// Bot should find the winning move at column 0
	winMove := bot.FindWinningMove(&board, models.PlayerColorRed)
	assert.Equal(t, 0, winMove)
}

func TestFindWinningMove_Diagonal(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Set up a diagonal winning opportunity
	// Row 0: R _ _ _
	// Row 1: Y R _ _
	// Row 2: Y Y R _
	// Row 3: _ _ _ _ (winning move here at col 3)
	board.MakeMove(0, models.PlayerColorRed)    // (0,0)
	board.MakeMove(1, models.PlayerColorYellow) // (0,1)
	board.MakeMove(1, models.PlayerColorRed)    // (1,1)
	board.MakeMove(2, models.PlayerColorYellow) // (0,2)
	board.MakeMove(2, models.PlayerColorYellow) // (1,2)
	board.MakeMove(2, models.PlayerColorRed)    // (2,2)
	board.MakeMove(3, models.PlayerColorYellow) // (0,3)
	board.MakeMove(3, models.PlayerColorYellow) // (1,3)
	board.MakeMove(3, models.PlayerColorYellow) // (2,3)

	// Bot should find the winning move at column 3 (row 3)
	winMove := bot.FindWinningMove(&board, models.PlayerColorRed)
	assert.Equal(t, 3, winMove)
}

func TestFindWinningMove_NoWin(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Only two pieces, no winning move
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(1, models.PlayerColorRed)

	winMove := bot.FindWinningMove(&board, models.PlayerColorRed)
	assert.Equal(t, -1, winMove)
}

func TestFindBlockingMove(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Set up opponent's winning threat: Y Y Y _
	board.MakeMove(0, models.PlayerColorYellow)
	board.MakeMove(1, models.PlayerColorYellow)
	board.MakeMove(2, models.PlayerColorYellow)

	// Bot (red) should block at column 3
	blockMove := bot.FindBlockingMove(&board, models.PlayerColorRed)
	assert.Equal(t, 3, blockMove)
}

func TestGetBestMove_TakesWinOverBlock(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Set up: Red can win at column 3, Yellow can win at column 6
	// Red: R R R _ at row 0
	// Yellow: Y Y Y _ at row 1 (stacked on different columns)
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(4, models.PlayerColorYellow)
	board.MakeMove(1, models.PlayerColorRed)
	board.MakeMove(5, models.PlayerColorYellow)
	board.MakeMove(2, models.PlayerColorRed)
	board.MakeMove(6, models.PlayerColorYellow)

	// Bot should take the win at column 3, not block at column 6
	bestMove := bot.GetBestMove(&board, models.PlayerColorRed, 4)
	assert.Equal(t, 3, bestMove, "Bot should take winning move over blocking")
}

func TestGetBestMove_BlocksWhenNoWin(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// Yellow has a winning threat, Red has no immediate win
	board.MakeMove(0, models.PlayerColorYellow)
	board.MakeMove(1, models.PlayerColorYellow)
	board.MakeMove(2, models.PlayerColorYellow)
	// Red has only one piece
	board.MakeMove(6, models.PlayerColorRed)

	// Bot should block at column 3
	bestMove := bot.GetBestMove(&board, models.PlayerColorRed, 4)
	assert.Equal(t, 3, bestMove, "Bot should block opponent's winning move")
}

func TestGetBestMove_PrefersCenterOnEmptyBoard(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	// On empty board, center column (3) is strategically best
	bestMove := bot.GetBestMove(&board, models.PlayerColorRed, 4)
	assert.Equal(t, 3, bestMove, "Bot should prefer center column on empty board")
}

func TestGetBestMoveWithTimeout(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	ctx := context.Background()
	timeout := 800 * time.Millisecond

	start := time.Now()
	move, err := bot.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, timeout)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.True(t, board.IsValidMove(move), "Move should be valid")
	assert.Less(t, elapsed, 1*time.Second, "Should complete within 1 second")
}

func TestGetBestMoveWithTimeout_RespectsContext(t *testing.T) {
	bot := NewMinimaxBot()
	board := models.NewBoard()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	move, err := bot.GetBestMoveWithTimeout(ctx, &board, models.PlayerColorRed, 1*time.Second)
	
	// Should still return a valid move (the default center)
	assert.True(t, board.IsValidMove(move) || move == 3)
	// Error might be context.Canceled
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}
}

func TestEvaluatePosition_EmptyBoard(t *testing.T) {
	bot := NewMinimaxBot().(*minimaxBot)
	board := models.NewBoard()

	score := bot.EvaluatePosition(&board, models.PlayerColorRed)
	assert.Equal(t, 0, score, "Empty board should have neutral score")
}

func TestEvaluatePosition_CenterAdvantage(t *testing.T) {
	bot := NewMinimaxBot().(*minimaxBot)
	
	// Board with red in center
	boardCenter := models.NewBoard()
	boardCenter.MakeMove(3, models.PlayerColorRed)
	
	// Board with red on edge
	boardEdge := models.NewBoard()
	boardEdge.MakeMove(0, models.PlayerColorRed)
	
	scoreCenter := bot.EvaluatePosition(&boardCenter, models.PlayerColorRed)
	scoreEdge := bot.EvaluatePosition(&boardEdge, models.PlayerColorRed)
	
	assert.Greater(t, scoreCenter, scoreEdge, "Center position should score higher")
}

func TestBotNeverMakesInvalidMove(t *testing.T) {
	bot := NewMinimaxBot()
	
	// Fill column 3 completely
	board := models.NewBoard()
	for i := 0; i < 6; i++ {
		if i%2 == 0 {
			board.MakeMove(3, models.PlayerColorRed)
		} else {
			board.MakeMove(3, models.PlayerColorYellow)
		}
	}
	
	// Bot should not choose column 3
	move := bot.GetBestMove(&board, models.PlayerColorRed, 4)
	assert.NotEqual(t, 3, move, "Bot should not choose full column")
	assert.True(t, board.IsValidMove(move), "Bot move should be valid")
}

func TestCopyBoard(t *testing.T) {
	original := models.NewBoard()
	original.MakeMove(3, models.PlayerColorRed)
	original.MakeMove(3, models.PlayerColorYellow)
	
	copied := copyBoard(&original)
	
	// Verify copy is equal
	assert.Equal(t, original.Grid, copied.Grid)
	assert.Equal(t, original.Height, copied.Height)
	
	// Modify copy and verify original is unchanged
	copied.MakeMove(0, models.PlayerColorRed)
	assert.NotEqual(t, original.Height[0], copied.Height[0])
}

func TestGetOpponent(t *testing.T) {
	assert.Equal(t, models.PlayerColorYellow, getOpponent(models.PlayerColorRed))
	assert.Equal(t, models.PlayerColorRed, getOpponent(models.PlayerColorYellow))
}
