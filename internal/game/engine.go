package game

import (
	"context"
	"fmt"

	"connect4-multiplayer/pkg/models"
	"connect4-multiplayer/internal/database/repositories"
)

// Engine defines the interface for the Connect 4 game engine
type Engine interface {
	// Game state operations
	CreateGame(ctx context.Context, player1, player2 string) (*models.GameSession, error)
	GetGame(ctx context.Context, gameID string) (*models.GameSession, error)
	
	// Move operations
	MakeMove(ctx context.Context, gameID string, playerUsername string, column int) (*MoveResult, error)
	ValidateMove(ctx context.Context, gameID string, playerUsername string, column int) error
	
	// Game state checks
	CheckGameEnd(ctx context.Context, game *models.GameSession) (*GameEndResult, error)
	IsPlayerTurn(ctx context.Context, game *models.GameSession, playerUsername string) bool
}

// MoveResult represents the result of making a move
type MoveResult struct {
	Move        *models.Move        `json:"move"`
	GameSession *models.GameSession `json:"gameSession"`
	GameEnded   bool                `json:"gameEnded"`
	Winner      *models.PlayerColor `json:"winner,omitempty"`
	IsDraw      bool                `json:"isDraw"`
}

// GameEndResult represents the result of a game ending
type GameEndResult struct {
	GameEnded bool                `json:"gameEnded"`
	Winner    *models.PlayerColor `json:"winner,omitempty"`
	IsDraw    bool                `json:"isDraw"`
	Reason    string              `json:"reason"`
}

// engine implements the Engine interface
type engine struct {
	gameRepo repositories.GameSessionRepository
	moveRepo repositories.MoveRepository
}

// NewEngine creates a new game engine instance
func NewEngine(gameRepo repositories.GameSessionRepository, moveRepo repositories.MoveRepository) Engine {
	return &engine{
		gameRepo: gameRepo,
		moveRepo: moveRepo,
	}
}

// CreateGame creates a new Connect 4 game session
func (e *engine) CreateGame(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	if player1 == "" || player2 == "" {
		return nil, fmt.Errorf("player usernames cannot be empty")
	}
	
	if player1 == player2 {
		return nil, fmt.Errorf("players must have different usernames")
	}
	
	game := &models.GameSession{
		Player1:     player1,
		Player2:     player2,
		Status:      models.StatusInProgress,
		CurrentTurn: models.PlayerColorRed, // Player1 always starts as red
		Board:       models.NewBoard(),
	}
	
	if err := e.gameRepo.Create(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	
	return game, nil
}

// GetGame retrieves a game session by ID
func (e *engine) GetGame(ctx context.Context, gameID string) (*models.GameSession, error) {
	game, err := e.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	return game, nil
}

// MakeMove processes a player's move
func (e *engine) MakeMove(ctx context.Context, gameID string, playerUsername string, column int) (*MoveResult, error) {
	// Get the current game state
	game, err := e.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	
	// Validate the move
	if err := e.ValidateMove(ctx, gameID, playerUsername, column); err != nil {
		return nil, err
	}
	
	// Get player color
	playerColor := game.GetPlayerColor(playerUsername)
	
	// Calculate the row where the disc will land
	row := game.Board.Height[column]
	
	// Make the move on the board
	if err := game.Board.MakeMove(column, playerColor); err != nil {
		return nil, fmt.Errorf("failed to make move on board: %w", err)
	}
	
	// Create move record
	move := &models.Move{
		GameID:  gameID,
		Player:  playerColor,
		Column:  column,
		Row:     row,
	}
	
	if err := e.moveRepo.Create(ctx, move); err != nil {
		return nil, fmt.Errorf("failed to save move: %w", err)
	}
	
	// Switch turns
	if game.CurrentTurn == models.PlayerColorRed {
		game.CurrentTurn = models.PlayerColorYellow
	} else {
		game.CurrentTurn = models.PlayerColorRed
	}
	
	// Check if game has ended
	gameEndResult, err := e.CheckGameEnd(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to check game end: %w", err)
	}
	
	// Update game status if ended
	if gameEndResult.GameEnded {
		if gameEndResult.IsDraw {
			game.Status = models.StatusCompleted
			game.Winner = nil
		} else if gameEndResult.Winner != nil {
			game.Status = models.StatusCompleted
			game.Winner = gameEndResult.Winner
		}
	}
	
	// Update the game session
	if err := e.gameRepo.Update(ctx, game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}
	
	return &MoveResult{
		Move:        move,
		GameSession: game,
		GameEnded:   gameEndResult.GameEnded,
		Winner:      gameEndResult.Winner,
		IsDraw:      gameEndResult.IsDraw,
	}, nil
}

// ValidateMove validates if a move is legal
func (e *engine) ValidateMove(ctx context.Context, gameID string, playerUsername string, column int) error {
	game, err := e.GetGame(ctx, gameID)
	if err != nil {
		return err
	}
	
	// Check if game is active
	if !game.IsActive() {
		return fmt.Errorf("game is not active (status: %s)", game.Status)
	}
	
	// Check if it's the player's turn
	if !e.IsPlayerTurn(ctx, game, playerUsername) {
		return fmt.Errorf("it's not %s's turn", playerUsername)
	}
	
	// Check if column is valid
	if column < 0 || column >= 7 {
		return fmt.Errorf("invalid column: %d (must be 0-6)", column)
	}
	
	// Check if column is not full
	if !game.Board.IsValidMove(column) {
		return fmt.Errorf("column %d is full", column)
	}
	
	return nil
}

// CheckGameEnd checks if the game has ended (win or draw)
func (e *engine) CheckGameEnd(ctx context.Context, game *models.GameSession) (*GameEndResult, error) {
	// Check for winner
	winner := game.Board.CheckWin()
	if winner != nil {
		return &GameEndResult{
			GameEnded: true,
			Winner:    winner,
			IsDraw:    false,
			Reason:    "four_in_a_row",
		}, nil
	}
	
	// Check for draw (board full)
	if game.Board.IsFull() {
		return &GameEndResult{
			GameEnded: true,
			Winner:    nil,
			IsDraw:    true,
			Reason:    "board_full",
		}, nil
	}
	
	// Game continues
	return &GameEndResult{
		GameEnded: false,
		Winner:    nil,
		IsDraw:    false,
		Reason:    "game_in_progress",
	}, nil
}

// IsPlayerTurn checks if it's the specified player's turn
func (e *engine) IsPlayerTurn(ctx context.Context, game *models.GameSession, playerUsername string) bool {
	currentPlayer := game.GetCurrentPlayer()
	return currentPlayer == playerUsername
}