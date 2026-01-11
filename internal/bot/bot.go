package bot

import (
	"context"
	"time"

	"connect4-multiplayer/pkg/models"
)

// Difficulty represents the bot difficulty level
type Difficulty int

const (
	DifficultyEasy   Difficulty = 1
	DifficultyMedium Difficulty = 2
	DifficultyHard   Difficulty = 3
)

// SearchDepth returns the search depth for a given difficulty
func (d Difficulty) SearchDepth() int {
	switch d {
	case DifficultyEasy:
		return 2
	case DifficultyMedium:
		return 4
	case DifficultyHard:
		return 7
	default:
		return 4
	}
}

// HumanDelay returns the simulated human-like delay for a given difficulty
func (d Difficulty) HumanDelay() time.Duration {
	switch d {
	case DifficultyEasy:
		return 500 * time.Millisecond
	case DifficultyMedium:
		return 300 * time.Millisecond
	case DifficultyHard:
		return 100 * time.Millisecond
	default:
		return 300 * time.Millisecond
	}
}

// MoveCalculator calculates the best move for a given board state
type MoveCalculator interface {
	GetBestMove(board *models.Board, player models.PlayerColor, depth int) int
}

// PositionEvaluator evaluates a board position for a given player
type PositionEvaluator interface {
	EvaluatePosition(board *models.Board, player models.PlayerColor) int
}

// BotAI combines move calculation and position evaluation
type BotAI interface {
	MoveCalculator
	PositionEvaluator
	// GetBestMoveWithTimeout returns the best move within the time limit
	GetBestMoveWithTimeout(ctx context.Context, board *models.Board, player models.PlayerColor, timeout time.Duration) (int, error)
	// FindWinningMove returns a winning move if one exists, -1 otherwise
	FindWinningMove(board *models.Board, player models.PlayerColor) int
	// FindBlockingMove returns a move that blocks opponent's win, -1 otherwise
	FindBlockingMove(board *models.Board, player models.PlayerColor) int
}
