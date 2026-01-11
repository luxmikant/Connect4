package bot

import (
	"context"
	"testing"
	"time"

	"connect4-multiplayer/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBotPlayerService(t *testing.T) {
	service := NewBotPlayerService()
	assert.NotNil(t, service)
}

func TestCreateBot_Easy(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyEasy)

	assert.NotNil(t, bot)
	assert.Equal(t, DifficultyEasy, bot.Difficulty)
	assert.Contains(t, bot.Username, BotUsernamePrefix)
	assert.Contains(t, bot.Username, "Easy")
	assert.NotNil(t, bot.AI)
}

func TestCreateBot_Medium(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)

	assert.NotNil(t, bot)
	assert.Equal(t, DifficultyMedium, bot.Difficulty)
	assert.Contains(t, bot.Username, BotUsernamePrefix)
	assert.Contains(t, bot.Username, "Medium")
	assert.NotNil(t, bot.AI)
}

func TestCreateBot_Hard(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyHard)

	assert.NotNil(t, bot)
	assert.Equal(t, DifficultyHard, bot.Difficulty)
	assert.Contains(t, bot.Username, BotUsernamePrefix)
	assert.Contains(t, bot.Username, "Hard")
	assert.NotNil(t, bot.AI)
}

func TestCreateBot_UniqueIDs(t *testing.T) {
	service := NewBotPlayerService()
	
	bot1 := service.CreateBot(DifficultyMedium)
	bot2 := service.CreateBot(DifficultyMedium)
	bot3 := service.CreateBot(DifficultyHard)

	assert.NotEqual(t, bot1.ID, bot2.ID)
	assert.NotEqual(t, bot2.ID, bot3.ID)
	assert.NotEqual(t, bot1.Username, bot2.Username)
}

func TestIsBot(t *testing.T) {
	service := NewBotPlayerService()

	// Bot usernames should be detected
	bot := service.CreateBot(DifficultyMedium)
	assert.True(t, service.IsBot(bot.Username))

	// Regular usernames should not be detected as bots
	assert.False(t, service.IsBot("player1"))
	assert.False(t, service.IsBot("john_doe"))
	assert.False(t, service.IsBot(""))
	assert.False(t, service.IsBot("Bot")) // Too short, missing underscore
}

func TestDifficulty_SearchDepth(t *testing.T) {
	assert.Equal(t, 2, DifficultyEasy.SearchDepth())
	assert.Equal(t, 4, DifficultyMedium.SearchDepth())
	assert.Equal(t, 7, DifficultyHard.SearchDepth())
	
	// Default case
	var unknown Difficulty = 99
	assert.Equal(t, 4, unknown.SearchDepth())
}

func TestDifficulty_HumanDelay(t *testing.T) {
	assert.Equal(t, 500*time.Millisecond, DifficultyEasy.HumanDelay())
	assert.Equal(t, 300*time.Millisecond, DifficultyMedium.HumanDelay())
	assert.Equal(t, 100*time.Millisecond, DifficultyHard.HumanDelay())
	
	// Default case
	var unknown Difficulty = 99
	assert.Equal(t, 300*time.Millisecond, unknown.HumanDelay())
}

func TestGetBotMove_ValidMove(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	board := models.NewBoard()

	ctx := context.Background()
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.True(t, board.IsValidMove(move), "Bot should return valid move")
}

func TestGetBotMove_TakesWinningMove(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	board := models.NewBoard()

	// Set up winning opportunity
	board.MakeMove(0, models.PlayerColorRed)
	board.MakeMove(1, models.PlayerColorRed)
	board.MakeMove(2, models.PlayerColorRed)

	ctx := context.Background()
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.Equal(t, 3, move, "Bot should take winning move")
}

func TestGetBotMove_BlocksOpponent(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	board := models.NewBoard()

	// Set up opponent's winning threat
	board.MakeMove(0, models.PlayerColorYellow)
	board.MakeMove(1, models.PlayerColorYellow)
	board.MakeMove(2, models.PlayerColorYellow)

	ctx := context.Background()
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.Equal(t, 3, move, "Bot should block opponent's winning move")
}

func TestGetBotMove_NilBot(t *testing.T) {
	service := NewBotPlayerService()
	board := models.NewBoard()

	ctx := context.Background()
	_, err := service.GetBotMove(ctx, nil, &board, models.PlayerColorRed)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGetBotMove_CompletesWithinTimeout(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyHard)
	board := models.NewBoard()

	ctx := context.Background()
	start := time.Now()
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.True(t, board.IsValidMove(move))
	assert.Less(t, elapsed, DefaultBotTimeout+100*time.Millisecond, "Should complete within timeout")
}

func TestGetBotMove_RespectsContextCancellation(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyHard)
	board := models.NewBoard()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	// Should still return a valid move or error
	if err == nil {
		assert.True(t, board.IsValidMove(move))
	}
}

func TestGetBotMove_HandlesFullBoard(t *testing.T) {
	service := NewBotPlayerService()
	bot := service.CreateBot(DifficultyMedium)
	
	// Create a nearly full board with one valid move
	board := models.NewBoard()
	
	// Fill all columns except column 3
	for col := 0; col < 7; col++ {
		if col == 3 {
			// Leave column 3 with 5 pieces (one spot left)
			for row := 0; row < 5; row++ {
				player := models.PlayerColorRed
				if row%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(col, player)
			}
		} else {
			// Fill other columns completely
			for row := 0; row < 6; row++ {
				player := models.PlayerColorRed
				if row%2 == 1 {
					player = models.PlayerColorYellow
				}
				board.MakeMove(col, player)
			}
		}
	}

	ctx := context.Background()
	move, err := service.GetBotMove(ctx, bot, &board, models.PlayerColorRed)

	require.NoError(t, err)
	assert.Equal(t, 3, move, "Bot should choose the only available column")
}
