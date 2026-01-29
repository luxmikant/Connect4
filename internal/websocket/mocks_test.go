//go:build property
// +build property

package websocket_test

import (
	"context"
	"sync"
	"time"

	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/pkg/models"
)

// MockMatchmakingService for testing
type MockMatchmakingService struct{}

func (m *MockMatchmakingService) JoinQueue(ctx context.Context, username string) (*matchmaking.QueueEntry, error) {
	return &matchmaking.QueueEntry{
		Username: username,
		JoinedAt: time.Now(),
		Timeout:  time.Now().Add(10 * time.Second),
	}, nil
}

func (m *MockMatchmakingService) LeaveQueue(ctx context.Context, username string) error {
	return nil
}

func (m *MockMatchmakingService) GetQueueStatus(ctx context.Context, username string) (*matchmaking.QueueStatus, error) {
	return &matchmaking.QueueStatus{
		InQueue:       false,
		Position:      0,
		WaitTime:      0,
		TimeRemaining: 0,
	}, nil
}

func (m *MockMatchmakingService) GetQueueLength(ctx context.Context) int {
	return 0
}

func (m *MockMatchmakingService) StartMatchmaking(ctx context.Context) error {
	return nil
}

func (m *MockMatchmakingService) StopMatchmaking() {
}

func (m *MockMatchmakingService) SetGameCreatedCallback(callback matchmaking.GameCreatedCallback) {
}

func (m *MockMatchmakingService) SetBotGameCallback(callback matchmaking.BotGameCallback) {
}

func (m *MockMatchmakingService) CreateBotGame(ctx context.Context, player string) (*models.GameSession, error) {
	return &models.GameSession{
		ID:          "bot-game-" + player,
		Player1:     player,
		Player2:     "bot_12345",
		Board:       models.NewBoard(),
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		StartTime:   time.Now(),
	}, nil
}

// MockGameService for testing
type MockGameService struct {
	sessions map[string]*models.GameSession
	mu       sync.RWMutex
}

func NewMockGameService() *MockGameService {
	return &MockGameService{
		sessions: make(map[string]*models.GameSession),
	}
}

func (m *MockGameService) CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &models.GameSession{
		ID:          generateTestID(),
		Player1:     player1,
		Player2:     player2,
		Board:       models.NewBoard(),
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		StartTime:   time.Now(),
	}

	m.sessions[session.ID] = session
	return session, nil
}

func (m *MockGameService) GetSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return nil, models.ErrGameNotFound
	}
	return session, nil
}

func (m *MockGameService) EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error {
	return m.CompleteGame(ctx, gameID, winner)
}

func (m *MockGameService) GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error) {
	session, err := m.GetSession(ctx, gameID)
	if err != nil {
		return "", "", err
	}
	return session.GetCurrentPlayer(), session.CurrentTurn, nil
}

func (m *MockGameService) SwitchTurn(ctx context.Context, gameID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}

	if session.CurrentTurn == models.PlayerColorRed {
		session.CurrentTurn = models.PlayerColorYellow
	} else {
		session.CurrentTurn = models.PlayerColorRed
	}

	return nil
}

func (m *MockGameService) AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error) {
	session, err := m.GetSession(ctx, gameID)
	if err != nil {
		return nil, err
	}

	colors := make(map[string]models.PlayerColor)
	colors[session.Player1] = models.PlayerColorRed
	colors[session.Player2] = models.PlayerColorYellow
	return colors, nil
}

func (m *MockGameService) CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}

	session.Status = models.StatusCompleted
	session.Winner = winner
	endTime := time.Now()
	session.EndTime = &endTime

	return nil
}

func (m *MockGameService) GetActiveSessions(ctx context.Context) ([]*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*models.GameSession
	for _, session := range m.sessions {
		if session.Status == models.StatusInProgress {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockGameService) GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []*models.GameSession
	for _, session := range m.sessions {
		if session.Player1 == username || session.Player2 == username {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *MockGameService) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	sessions, err := m.GetSessionsByPlayer(ctx, username)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.Status == models.StatusInProgress {
			return session, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameService) GetActiveSessionCount(ctx context.Context) (int64, error) {
	sessions, err := m.GetActiveSessions(ctx)
	if err != nil {
		return 0, err
	}
	return int64(len(sessions)), nil
}

func (m *MockGameService) CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error) {
	return 0, nil
}

func (m *MockGameService) MarkSessionAbandoned(ctx context.Context, gameID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[gameID]
	if !exists {
		return models.ErrGameNotFound
	}

	session.Status = models.StatusAbandoned
	return nil
}

func (m *MockGameService) StartCleanupWorker(ctx context.Context, interval time.Duration) {}

func (m *MockGameService) StopCleanupWorker() {}

func (m *MockGameService) MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) MarkPlayerReconnected(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) GetDisconnectedPlayers(gameID string) map[string]time.Time {
	return make(map[string]time.Time)
}

func (m *MockGameService) HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error {
	return nil
}

func (m *MockGameService) CacheSession(session *models.GameSession) {}

func (m *MockGameService) GetCachedSession(gameID string) (*models.GameSession, bool) {
	session, err := m.GetSession(context.Background(), gameID)
	if err != nil {
		return nil, false
	}
	return session, true
}

func (m *MockGameService) InvalidateCache(gameID string) {}

func (m *MockGameService) CleanupCache(maxAge time.Duration) int {
	return 0
}

func (m *MockGameService) GetCacheStats() map[string]interface{} {
	return make(map[string]interface{})
}

// Custom room methods
func (m *MockGameService) CreateCustomRoom(ctx context.Context, creator string) (*models.GameSession, string, error) {
	session := &models.GameSession{
		ID:          generateTestID(),
		Player1:     creator,
		Board:       models.NewBoard(),
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusWaiting,
		StartTime:   time.Now(),
		RoomCode:    stringPtr("TEST1234"),
	}
	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()
	return session, "TEST1234", nil
}

func (m *MockGameService) JoinCustomRoom(ctx context.Context, roomCode, username string) (*models.GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, session := range m.sessions {
		if session.RoomCode != nil && *session.RoomCode == roomCode && session.Status == models.StatusWaiting {
			session.Player2 = username
			session.Status = models.StatusInProgress
			return session, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameService) GetSessionByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, session := range m.sessions {
		if session.RoomCode != nil && *session.RoomCode == roomCode {
			return session, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameService) RematchCustomRoom(ctx context.Context, gameID, username string) (*models.GameSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldSession, exists := m.sessions[gameID]
	if !exists {
		return nil, models.ErrGameNotFound
	}
	// Create new session with same players
	newSession := &models.GameSession{
		ID:          generateTestID(),
		Player1:     oldSession.Player1,
		Player2:     oldSession.Player2,
		Board:       models.NewBoard(),
		CurrentTurn: models.PlayerColorRed,
		Status:      models.StatusInProgress,
		StartTime:   time.Now(),
		RoomCode:    oldSession.RoomCode,
	}
	// Clear room code from old session
	oldSession.RoomCode = nil
	m.sessions[newSession.ID] = newSession
	return newSession, nil
}

func stringPtr(s string) *string {
	return &s
}

func generateTestID() string {
	return "test-" + time.Now().Format("20060102150405")
}
