package bot

import (
	"context"
	"math"
	"time"

	"connect4-multiplayer/pkg/models"
)

const (
	// Scoring constants for position evaluation
	scoreWin         = 100000
	scoreLose        = -100000
	scoreThreeInRow  = 100
	scoreTwoInRow    = 10
	scoreCenterBonus = 3
)

// minimaxBot implements the BotAI interface using minimax with alpha-beta pruning
type minimaxBot struct {
	transpositionTable map[string]int // Cache for evaluated positions
}

// NewMinimaxBot creates a new minimax bot instance
func NewMinimaxBot() BotAI {
	return &minimaxBot{
		transpositionTable: make(map[string]int),
	}
}

// GetBestMove returns the best move using minimax with alpha-beta pruning
func (b *minimaxBot) GetBestMove(board *models.Board, player models.PlayerColor, depth int) int {
	// First, check for immediate winning move
	winMove := b.FindWinningMove(board, player)
	if winMove != -1 {
		return winMove
	}

	// Second, check if we need to block opponent's winning move
	opponent := getOpponent(player)
	blockMove := b.FindWinningMove(board, opponent)
	if blockMove != -1 {
		return blockMove
	}

	// Use minimax with alpha-beta pruning for strategic move
	bestMove := 3 // Default to center column
	bestScore := math.MinInt32

	// Order moves: center columns first for better pruning
	moveOrder := []int{3, 2, 4, 1, 5, 0, 6}

	for _, col := range moveOrder {
		if !board.IsValidMove(col) {
			continue
		}

		// Make move on a copy of the board
		boardCopy := copyBoard(board)
		if err := boardCopy.MakeMove(col, player); err != nil {
			continue // Skip invalid moves
		}

		// Evaluate using minimax
		score := b.minimax(boardCopy, depth-1, math.MinInt32, math.MaxInt32, false, player)

		if score > bestScore {
			bestScore = score
			bestMove = col
		}
	}

	return bestMove
}

// GetBestMoveWithTimeout returns the best move within the time limit using iterative deepening
func (b *minimaxBot) GetBestMoveWithTimeout(ctx context.Context, board *models.Board, player models.PlayerColor, timeout time.Duration) (int, error) {
	// Clear transposition table for new search
	b.transpositionTable = make(map[string]int)

	deadline := time.Now().Add(timeout)
	bestMove := 3 // Default to center

	// First check for immediate winning/blocking moves (these are fast)
	winMove := b.FindWinningMove(board, player)
	if winMove != -1 {
		return winMove, nil
	}

	opponent := getOpponent(player)
	blockMove := b.FindWinningMove(board, opponent)
	if blockMove != -1 {
		return blockMove, nil
	}

	// Find any valid move as fallback
	for col := 0; col < 7; col++ {
		if board.IsValidMove(col) {
			bestMove = col
			break
		}
	}

	// Iterative deepening: start shallow, go deeper if time permits
	for depth := 1; depth <= 7; depth++ {
		select {
		case <-ctx.Done():
			return bestMove, ctx.Err()
		default:
			if time.Now().After(deadline) {
				return bestMove, nil
			}

			// Use a limited search for this depth
			move := b.getBestMoveWithDeadline(board, player, depth, deadline)
			if board.IsValidMove(move) {
				bestMove = move
			}

			// Check time after each depth
			if time.Now().After(deadline) {
				return bestMove, nil
			}
		}
	}

	return bestMove, nil
}

// getBestMoveWithDeadline performs minimax search with a deadline check
func (b *minimaxBot) getBestMoveWithDeadline(board *models.Board, player models.PlayerColor, depth int, deadline time.Time) int {
	bestMove := 3 // Default to center column
	bestScore := math.MinInt32

	// Order moves: center columns first for better pruning
	moveOrder := []int{3, 2, 4, 1, 5, 0, 6}

	for _, col := range moveOrder {
		if time.Now().After(deadline) {
			break
		}

		if !board.IsValidMove(col) {
			continue
		}

		// Make move on a copy of the board
		boardCopy := copyBoard(board)
		if err := boardCopy.MakeMove(col, player); err != nil {
			continue // Skip invalid moves
		}

		// Evaluate using minimax
		score := b.minimaxWithDeadline(boardCopy, depth-1, math.MinInt32, math.MaxInt32, false, player, deadline)

		if score > bestScore {
			bestScore = score
			bestMove = col
		}
	}

	return bestMove
}

// minimaxWithDeadline implements minimax with deadline checking
func (b *minimaxBot) minimaxWithDeadline(board *models.Board, depth int, alpha, beta int, isMaximizing bool, botPlayer models.PlayerColor, deadline time.Time) int {
	// Check deadline
	if time.Now().After(deadline) {
		return b.EvaluatePosition(board, botPlayer)
	}

	// Check for terminal states
	winner := board.CheckWin()
	if winner != nil {
		if *winner == botPlayer {
			return scoreWin + depth // Prefer faster wins
		}
		return scoreLose - depth // Prefer slower losses
	}

	if board.IsFull() {
		return 0 // Draw
	}

	if depth == 0 {
		return b.EvaluatePosition(board, botPlayer)
	}

	moveOrder := []int{3, 2, 4, 1, 5, 0, 6}

	if isMaximizing {
		maxScore := math.MinInt32
		for _, col := range moveOrder {
			if time.Now().After(deadline) {
				break
			}

			if !board.IsValidMove(col) {
				continue
			}

			boardCopy := copyBoard(board)
			if err := boardCopy.MakeMove(col, botPlayer); err != nil {
				continue // Skip invalid moves
			}

			score := b.minimaxWithDeadline(boardCopy, depth-1, alpha, beta, false, botPlayer, deadline)
			maxScore = max(maxScore, score)
			alpha = max(alpha, score)

			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxScore
	} else {
		minScore := math.MaxInt32
		opponent := getOpponent(botPlayer)
		for _, col := range moveOrder {
			if time.Now().After(deadline) {
				break
			}

			if !board.IsValidMove(col) {
				continue
			}

			boardCopy := copyBoard(board)
			if err := boardCopy.MakeMove(col, opponent); err != nil {
				continue // Skip invalid moves
			}

			score := b.minimaxWithDeadline(boardCopy, depth-1, alpha, beta, true, botPlayer, deadline)
			minScore = min(minScore, score)
			beta = min(beta, score)

			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minScore
	}
}

// minimax implements the minimax algorithm with alpha-beta pruning
func (b *minimaxBot) minimax(board *models.Board, depth int, alpha, beta int, isMaximizing bool, botPlayer models.PlayerColor) int {
	// Check for terminal states
	winner := board.CheckWin()
	if winner != nil {
		if *winner == botPlayer {
			return scoreWin + depth // Prefer faster wins
		}
		return scoreLose - depth // Prefer slower losses
	}

	if board.IsFull() {
		return 0 // Draw
	}

	if depth == 0 {
		return b.EvaluatePosition(board, botPlayer)
	}

	moveOrder := []int{3, 2, 4, 1, 5, 0, 6}

	if isMaximizing {
		maxScore := math.MinInt32
		for _, col := range moveOrder {
			if !board.IsValidMove(col) {
				continue
			}

			boardCopy := copyBoard(board)
			if err := boardCopy.MakeMove(col, botPlayer); err != nil {
				continue // Skip invalid moves
			}

			score := b.minimax(boardCopy, depth-1, alpha, beta, false, botPlayer)
			maxScore = max(maxScore, score)
			alpha = max(alpha, score)

			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxScore
	} else {
		minScore := math.MaxInt32
		opponent := getOpponent(botPlayer)
		for _, col := range moveOrder {
			if !board.IsValidMove(col) {
				continue
			}

			boardCopy := copyBoard(board)
			if err := boardCopy.MakeMove(col, opponent); err != nil {
				continue // Skip invalid moves
			}

			score := b.minimax(boardCopy, depth-1, alpha, beta, true, botPlayer)
			minScore = min(minScore, score)
			beta = min(beta, score)

			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minScore
	}
}

// EvaluatePosition evaluates the board position for the given player
func (b *minimaxBot) EvaluatePosition(board *models.Board, player models.PlayerColor) int {
	score := 0
	opponent := getOpponent(player)

	// Evaluate all windows of 4
	score += b.evaluateWindows(board, player, opponent)

	// Center column bonus
	score += b.evaluateCenterControl(board, player)

	return score
}

// evaluateWindows evaluates all possible 4-cell windows on the board
func (b *minimaxBot) evaluateWindows(board *models.Board, player, opponent models.PlayerColor) int {
	score := 0

	// Horizontal windows
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			window := []models.PlayerColor{
				board.Grid[row][col],
				board.Grid[row][col+1],
				board.Grid[row][col+2],
				board.Grid[row][col+3],
			}
			score += b.evaluateWindow(window, player, opponent)
		}
	}

	// Vertical windows
	for row := 0; row < 3; row++ {
		for col := 0; col < 7; col++ {
			window := []models.PlayerColor{
				board.Grid[row][col],
				board.Grid[row+1][col],
				board.Grid[row+2][col],
				board.Grid[row+3][col],
			}
			score += b.evaluateWindow(window, player, opponent)
		}
	}

	// Diagonal windows (bottom-left to top-right)
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			window := []models.PlayerColor{
				board.Grid[row][col],
				board.Grid[row+1][col+1],
				board.Grid[row+2][col+2],
				board.Grid[row+3][col+3],
			}
			score += b.evaluateWindow(window, player, opponent)
		}
	}

	// Diagonal windows (top-left to bottom-right)
	for row := 3; row < 6; row++ {
		for col := 0; col < 4; col++ {
			window := []models.PlayerColor{
				board.Grid[row][col],
				board.Grid[row-1][col+1],
				board.Grid[row-2][col+2],
				board.Grid[row-3][col+3],
			}
			score += b.evaluateWindow(window, player, opponent)
		}
	}

	return score
}

// evaluateWindow evaluates a single 4-cell window
func (b *minimaxBot) evaluateWindow(window []models.PlayerColor, player, opponent models.PlayerColor) int {
	playerCount := 0
	opponentCount := 0
	emptyCount := 0

	for _, cell := range window {
		switch cell {
		case player:
			playerCount++
		case opponent:
			opponentCount++
		default:
			emptyCount++
		}
	}

	// If window has both player and opponent pieces, it's blocked
	if playerCount > 0 && opponentCount > 0 {
		return 0
	}

	// Score based on player pieces
	if playerCount == 4 {
		return scoreWin
	}
	if playerCount == 3 && emptyCount == 1 {
		return scoreThreeInRow
	}
	if playerCount == 2 && emptyCount == 2 {
		return scoreTwoInRow
	}

	// Penalize opponent threats
	if opponentCount == 4 {
		return scoreLose
	}
	if opponentCount == 3 && emptyCount == 1 {
		return -scoreThreeInRow
	}
	if opponentCount == 2 && emptyCount == 2 {
		return -scoreTwoInRow
	}

	return 0
}

// evaluateCenterControl gives bonus for controlling center columns
func (b *minimaxBot) evaluateCenterControl(board *models.Board, player models.PlayerColor) int {
	score := 0
	centerCol := 3

	for row := 0; row < 6; row++ {
		if board.Grid[row][centerCol] == player {
			score += scoreCenterBonus
		}
	}

	// Also give smaller bonus for adjacent center columns
	for row := 0; row < 6; row++ {
		if board.Grid[row][2] == player {
			score += scoreCenterBonus / 2
		}
		if board.Grid[row][4] == player {
			score += scoreCenterBonus / 2
		}
	}

	return score
}

// FindWinningMove finds a move that wins the game immediately
func (b *minimaxBot) FindWinningMove(board *models.Board, player models.PlayerColor) int {
	for col := 0; col < 7; col++ {
		if !board.IsValidMove(col) {
			continue
		}

		boardCopy := copyBoard(board)
		if err := boardCopy.MakeMove(col, player); err != nil {
			continue // Skip invalid moves
		}

		winner := boardCopy.CheckWin()
		if winner != nil && *winner == player {
			return col
		}
	}
	return -1
}

// FindBlockingMove finds a move that blocks the opponent's winning move
func (b *minimaxBot) FindBlockingMove(board *models.Board, player models.PlayerColor) int {
	opponent := getOpponent(player)
	return b.FindWinningMove(board, opponent)
}

// Helper functions

func getOpponent(player models.PlayerColor) models.PlayerColor {
	if player == models.PlayerColorRed {
		return models.PlayerColorYellow
	}
	return models.PlayerColorRed
}

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
