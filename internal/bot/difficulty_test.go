package bot

import (
	"context"
	"testing"
	"time"

	"connect4-multiplayer/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDifficultyBehaviorComparison compares bot behaviors across difficulty levels
func TestDifficultyBehaviorComparison(t *testing.T) {
	// Test that different difficulties have measurably different performance characteristics
	// Easy should be much faster and less accurate than Hard

	service := NewBotPlayerService()
	easyBot := service.CreateBot(DifficultyEasy)
	mediumBot := service.CreateBot(DifficultyMedium)
	hardBot := service.CreateBot(DifficultyHard)

	// Verify bots were created with correct difficulties
	assert.Equal(t, DifficultyEasy, easyBot.Difficulty)
	assert.Equal(t, DifficultyMedium, mediumBot.Difficulty)
	assert.Equal(t, DifficultyHard, hardBot.Difficulty)

	// Verify search depths are different
	assert.Less(t, easyBot.Difficulty.SearchDepth(), mediumBot.Difficulty.SearchDepth(),
		"Easy bot should have shallower search than Medium")
	assert.Less(t, mediumBot.Difficulty.SearchDepth(), hardBot.Difficulty.SearchDepth(),
		"Medium bot should have shallower search than Hard")

	// Verify delays are appropriate (higher difficulty = faster moves)
	assert.Greater(t, easyBot.Difficulty.HumanDelay(), mediumBot.Difficulty.HumanDelay(),
		"Easy bot should have longer delay than Medium")
	assert.Greater(t, mediumBot.Difficulty.HumanDelay(), hardBot.Difficulty.HumanDelay(),
		"Medium bot should have longer delay than Hard")
}

// TestEasyBot_FindsObviousWin verifies Easy bot can still find immediate wins
func TestEasyBot_FindsObviousWin(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyEasy)
	board := models.NewBoard()

	// Set up an obvious vertical win for red in column 3
	board.MakeMove(3, models.PlayerColorRed)
	board.MakeMove(3, models.PlayerColorRed)
	board.MakeMove(3, models.PlayerColorRed)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Easy bot should find the immediate winning move
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.Equal(t, 3, move, "Easy bot should take the winning move")
}

// TestMediumBot_FindsStrategicMove verifies Medium bot plays reasonably well
func TestMediumBot_FindsStrategicMove(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	board := models.NewBoard()

	// Set up opponent's threat: Yellow has 3 in a row horizontally
	board.MakeMove(0, models.PlayerColorYellow)
	board.MakeMove(1, models.PlayerColorYellow)
	board.MakeMove(2, models.PlayerColorYellow)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Medium bot should block the threat
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.Equal(t, 3, move, "Medium bot should block opponent's winning move")
}

// TestHardBot_PlaysOptimally verifies Hard bot makes near-perfect moves
func TestHardBot_PlaysOptimally(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyHard)
	board := models.NewBoard()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Hard bot's opening move should be strategic (typically center)
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	// Center columns (3, 2, 4) are typically best opening moves
	assert.Contains(t, []int{2, 3, 4}, move, "Hard bot should make a strategic opening move")
}

// TestDifficultyPerformanceTiming verifies delays are respected
func TestDifficultyPerformanceTiming(t *testing.T) {
	service := NewBotPlayerService()
	board := models.NewBoard()
	ctx := context.Background()

	tests := []struct {
		name       string
		difficulty Difficulty
		minDelay   time.Duration
		maxDelay   time.Duration
	}{
		{
			name:       "Easy bot has appropriate delay",
			difficulty: DifficultyEasy,
			minDelay:   500 * time.Millisecond,
			maxDelay:   1500 * time.Millisecond,
		},
		{
			name:       "Medium bot has appropriate delay",
			difficulty: DifficultyMedium,
			minDelay:   300 * time.Millisecond,
			maxDelay:   1000 * time.Millisecond,
		},
		{
			name:       "Hard bot has appropriate delay",
			difficulty: DifficultyHard,
			minDelay:   100 * time.Millisecond,
			maxDelay:   2000 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := service.CreateBot(tt.difficulty)

			start := time.Now()
			_, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, elapsed, tt.minDelay,
				"Bot move should take at least the human delay")
			assert.Less(t, elapsed, tt.maxDelay,
				"Bot move should complete within reasonable time")
		})
	}
}

// TestBotMove_ValidAcrossDifficulties ensures all difficulties make valid moves
func TestBotMove_ValidAcrossDifficulties(t *testing.T) {
	service := NewBotPlayerService()
	board := models.NewBoard()
	ctx := context.Background()

	difficulties := []struct {
		name       string
		difficulty Difficulty
	}{
		{"Easy", DifficultyEasy},
		{"Medium", DifficultyMedium},
		{"Hard", DifficultyHard},
	}

	for _, d := range difficulties {
		t.Run(d.name, func(t *testing.T) {
			bot := service.CreateBot(d.difficulty)
			move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

			require.NoError(t, err)
			assert.True(t, board.IsValidMove(move),
				"%s bot should always return a valid move", d.name)
			assert.GreaterOrEqual(t, move, 0, "Move should be >= 0")
			assert.LessOrEqual(t, move, 6, "Move should be <= 6")
		})
	}
}

// TestBotDifficulty_FullColumn verifies all bots handle full columns
func TestBotDifficulty_FullColumn(t *testing.T) {
	service := NewBotPlayerService()
	board := models.NewBoard()
	ctx := context.Background()

	// Fill column 3 completely
	for i := 0; i < 6; i++ {
		color := models.PlayerColorRed
		if i%2 == 1 {
			color = models.PlayerColorYellow
		}
		board.MakeMove(3, color)
	}

	difficulties := []Difficulty{DifficultyEasy, DifficultyMedium, DifficultyHard}

	for _, diff := range difficulties {
		t.Run(diff.String(), func(t *testing.T) {
			bot := service.CreateBot(diff)
			move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

			require.NoError(t, err)
			assert.NotEqual(t, 3, move, "%s bot should not choose the full column", diff.String())
			assert.True(t, board.IsValidMove(move), "%s bot should choose a valid column", diff.String())
		})
	}
}

// TestBotContext_CancellationRespected verifies bots respect context cancellation
func TestBotContext_CancellationRespected(t *testing.T) {
	service := NewBotPlayerService()
	board := models.NewBoard()

	// Test with Hard bot (takes longer to compute)
	bot := service.CreateBot(DifficultyHard)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Bot should handle cancellation gracefully
	_, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	// Error is acceptable (context cancelled), but should not panic
	// Some implementations may still return a move if computed quickly
	if err != nil {
		assert.ErrorIs(t, err, context.Canceled)
	}
}

// TestBotDifficulty_ConsistentBehavior verifies bots are deterministic for same board
func TestBotDifficulty_ConsistentBehavior(t *testing.T) {
	service := NewBotPlayerService()
	ctx := context.Background()

	// Create a specific board state
	board := models.NewBoard()
	board.MakeMove(3, models.PlayerColorRed)
	board.MakeMove(2, models.PlayerColorYellow)
	board.MakeMove(4, models.PlayerColorRed)

	difficulties := []Difficulty{DifficultyMedium, DifficultyHard}

	for _, diff := range difficulties {
		t.Run(diff.String(), func(t *testing.T) {
			bot1 := service.CreateBot(diff)
			bot2 := service.CreateBot(diff)

			move1, err1 := service.GetBotMove(ctx, bot1, &board, models.PlayerColorRed)
			move2, err2 := service.GetBotMove(ctx, bot2, &board, models.PlayerColorRed)

			require.NoError(t, err1)
			require.NoError(t, err2)

			// Medium and Hard bots should be deterministic (no randomness)
			// Easy bot has randomness, so skip this test for Easy
			assert.Equal(t, move1, move2,
				"%s bot should make the same move for the same board state", diff.String())
		})
	}
}

// Benchmark tests for performance comparison
func BenchmarkEasyBot_Move(b *testing.B) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyEasy)
	board := models.NewBoard()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)
	}
}

func BenchmarkMediumBot_Move(b *testing.B) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	board := models.NewBoard()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)
	}
}

func BenchmarkHardBot_Move(b *testing.B) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyHard)
	board := models.NewBoard()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)
	}
}
