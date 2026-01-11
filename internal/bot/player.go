package bot

import (
	"context"
	"fmt"
	"time"

	"connect4-multiplayer/pkg/models"
)

const (
	// BotUsernamePrefix is the prefix for bot usernames
	BotUsernamePrefix = "Bot_"
	// DefaultBotTimeout is the maximum time allowed for bot to make a move
	DefaultBotTimeout = 1 * time.Second
)

// BotPlayer represents a bot player in the game
type BotPlayer struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Difficulty Difficulty `json:"difficulty"`
	AI         BotAI      `json:"-"`
}

// BotPlayerService manages bot players
type BotPlayerService interface {
	// CreateBot creates a new bot player with the specified difficulty
	CreateBot(difficulty Difficulty) *BotPlayer
	// GetBotMove gets the best move for the bot given the current board state
	GetBotMove(ctx context.Context, bot *BotPlayer, board *models.Board, color models.PlayerColor) (int, error)
	// IsBot checks if a username belongs to a bot
	IsBot(username string) bool
}

// botPlayerService implements BotPlayerService
type botPlayerService struct {
	botCounter int
}

// NewBotPlayerService creates a new bot player service
func NewBotPlayerService() BotPlayerService {
	return &botPlayerService{
		botCounter: 0,
	}
}

// CreateBot creates a new bot player with the specified difficulty
func (s *botPlayerService) CreateBot(difficulty Difficulty) *BotPlayer {
	s.botCounter++
	
	difficultyName := "Medium"
	switch difficulty {
	case DifficultyEasy:
		difficultyName = "Easy"
	case DifficultyHard:
		difficultyName = "Hard"
	}
	
	return &BotPlayer{
		ID:         fmt.Sprintf("bot_%d", s.botCounter),
		Username:   fmt.Sprintf("%s%s_%d", BotUsernamePrefix, difficultyName, s.botCounter),
		Difficulty: difficulty,
		AI:         NewMinimaxBot(),
	}
}

// GetBotMove gets the best move for the bot given the current board state
func (s *botPlayerService) GetBotMove(ctx context.Context, bot *BotPlayer, board *models.Board, color models.PlayerColor) (int, error) {
	if bot == nil {
		return -1, fmt.Errorf("bot player is nil")
	}
	
	if bot.AI == nil {
		return -1, fmt.Errorf("bot AI is not initialized")
	}
	
	// Add human-like delay based on difficulty
	delay := bot.Difficulty.HumanDelay()
	
	// Calculate timeout (must complete within 1 second total including delay)
	timeout := DefaultBotTimeout - delay
	if timeout < 100*time.Millisecond {
		timeout = 100 * time.Millisecond
	}
	
	// Create a context with timeout
	moveCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Start timing
	start := time.Now()
	
	// Get the best move
	move, err := bot.AI.GetBestMoveWithTimeout(moveCtx, board, color, timeout)
	if err != nil && err != context.DeadlineExceeded {
		return -1, fmt.Errorf("failed to get bot move: %w", err)
	}
	
	// Calculate remaining delay time
	elapsed := time.Since(start)
	remainingDelay := delay - elapsed
	if remainingDelay > 0 {
		select {
		case <-ctx.Done():
			return move, ctx.Err()
		case <-time.After(remainingDelay):
			// Delay completed
		}
	}
	
	// Validate the move
	if !board.IsValidMove(move) {
		// Fallback: find any valid move
		for col := 0; col < 7; col++ {
			if board.IsValidMove(col) {
				return col, nil
			}
		}
		return -1, fmt.Errorf("no valid moves available")
	}
	
	return move, nil
}

// IsBot checks if a username belongs to a bot
func (s *botPlayerService) IsBot(username string) bool {
	if len(username) < len(BotUsernamePrefix) {
		return false
	}
	return username[:len(BotUsernamePrefix)] == BotUsernamePrefix
}
