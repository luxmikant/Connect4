package game_test

import (
	"context"
	"time"

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
	if game, exists := m.games[id]; exists {
		return game, nil
	}
	return nil, models.ErrGameNotFound
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
	var activeGames []*models.GameSession
	for _, game := range m.games {
		if game.IsActive() {
			activeGames = append(activeGames, game)
		}
	}
	return activeGames, nil
}

func (m *MockGameSessionRepository) GetGamesByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	var playerGames []*models.GameSession
	for _, game := range m.games {
		if game.Player1 == username || game.Player2 == username {
			playerGames = append(playerGames, game)
		}
	}
	return playerGames, nil
}

func (m *MockGameSessionRepository) GetGameHistory(ctx context.Context, limit, offset int) ([]*models.GameSession, error) {
	var history []*models.GameSession
	for _, game := range m.games {
		history = append(history, game)
	}
	return history, nil
}

func (m *MockGameSessionRepository) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	for _, game := range m.games {
		if game.IsActive() && (game.Player1 == username || game.Player2 == username) {
			return game, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameSessionRepository) GetActiveSessionCount(ctx context.Context) (int64, error) {
	count := int64(0)
	for _, game := range m.games {
		if game.IsActive() {
			count++
		}
	}
	return count, nil
}

func (m *MockGameSessionRepository) GetTimedOutSessions(ctx context.Context, timeout time.Duration) ([]*models.GameSession, error) {
	var timedOut []*models.GameSession
	cutoff := time.Now().Add(-timeout)
	for _, game := range m.games {
		if game.UpdatedAt.Before(cutoff) {
			timedOut = append(timedOut, game)
		}
	}
	return timedOut, nil
}

func (m *MockGameSessionRepository) BulkUpdateStatus(ctx context.Context, sessionIDs []string, status models.GameStatus) error {
	for _, id := range sessionIDs {
		if game, exists := m.games[id]; exists {
			game.Status = status
		}
	}
	return nil
}

func (m *MockGameSessionRepository) GetByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	for _, game := range m.games {
		if game.RoomCode != nil && *game.RoomCode == roomCode {
			return game, nil
		}
	}
	return nil, nil
}

// MockMoveRepository for testing
type MockMoveRepository struct {
	moves map[string][]*models.Move
}

func NewMockMoveRepository() *MockMoveRepository {
	return &MockMoveRepository{
		moves: make(map[string][]*models.Move),
	}
}

func (m *MockMoveRepository) Create(ctx context.Context, move *models.Move) error {
	if move.ID == "" {
		move.ID = "test-move-id"
	}
	m.moves[move.GameID] = append(m.moves[move.GameID], move)
	return nil
}

func (m *MockMoveRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.Move, error) {
	return m.moves[gameID], nil
}

func (m *MockMoveRepository) GetByID(ctx context.Context, id string) (*models.Move, error) {
	for _, gameMoves := range m.moves {
		for _, move := range gameMoves {
			if move.ID == id {
				return move, nil
			}
		}
	}
	return nil, models.ErrMoveNotFound
}

func (m *MockMoveRepository) Delete(ctx context.Context, id string) error {
	for gameID, gameMoves := range m.moves {
		for i, move := range gameMoves {
			if move.ID == id {
				m.moves[gameID] = append(gameMoves[:i], gameMoves[i+1:]...)
				return nil
			}
		}
	}
	return models.ErrMoveNotFound
}

func (m *MockMoveRepository) GetMoveHistory(ctx context.Context, gameID string, limit int) ([]*models.Move, error) {
	moves := m.moves[gameID]
	if limit > 0 && len(moves) > limit {
		return moves[:limit], nil
	}
	return moves, nil
}
