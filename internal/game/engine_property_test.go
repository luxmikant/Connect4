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

// MockGameSessionRepository for testing
type MockGameSessionRepository struct {
	games map[string]*models.GameSession
}

func NewMockGameSessionRepository() *MockGameSessionRepository {
	return &MockGameSessionRepository{
		games: make(map[string]*models.GameSession),
	}
}

func (m *MockGameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	if session.ID == "" {
		session.ID = "test-game-id"
	}
	m.games[session.ID] = session
	return nil
}

func (m *MockGameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	game, exists := m.games[id]
	if !exists {
		return nil, models.ErrGameNotFound
	}
	return game, nil
}

func (m *MockGameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	m.games[session.ID] = session
	return nil
}

func (m *MockGameSessionRepository) Delete(ctx context.Context, id string) error {
	delete(m.games, id)
	return nil
}

func (m *MockGameSessionRepository) GetActiveGames(ctx context.Context) ([]*models.GameSession, error) {
	var active []*models.GameSession
	for _, game := range m.games {
		if game.IsActive() {
			active = append(active, game)
		}
	}
	return active, nil
}

func (m *MockGameSessionRepository) GetGamesByPlayer(ctx context.Context, playerID string) ([]*models.GameSession, error) {
	var games []*models.GameSession
	for _, game := range m.games {
		if game.Player1 == playerID || game.Player2 == playerID {
			games = append(games, game)
		}
	}
	return games, nil
}

func (m *MockGameSessionRepository) GetGameHistory(ctx context.Context, limit, offset int) ([]*models.GameSession, error) {
	var history []*models.GameSession
	for _, game := range m.games {
		history = append(history, game)
	}
	return history, nil
}

// MockMoveRepository for testing
type MockMoveRepository struct {
	moves map[string]*models.Move
}

func NewMockMoveRepository() *MockMoveRepository {
	return &MockMoveRepository{
		moves: make(map[string]*models.Move),
	}
}

func (m *MockMoveRepository) Create(ctx context.Context, move *models.Move) error {
	if move.ID == "" {
		move.ID = "test-move-id"
	}
	m.moves[move.ID] = move
	return nil
}

func (m *MockMoveRepository) GetByID(ctx context.Context, id string) (*models.Move, error) {
	move, exists := m.moves[id]
	if !exists {
		return nil, models.ErrMoveNotFound
	}
	return move, nil
}

func (m *MockMoveRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.Move, error) {
	var gameMoves []*models.Move
	for _, move := range m.moves {
		if move.GameID == gameID {
			gameMoves = append(gameMoves, move)
		}
	}
	return gameMoves, nil
}

func (m *MockMoveRepository) Delete(ctx context.Context, id string) error {
	delete(m.moves, id)
	return nil
}

func (m *MockMoveRepository) GetMoveHistory(ctx context.Context, gameID string, limit int) ([]*models.Move, error) {
	return m.GetByGameID(ctx, gameID)
}

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