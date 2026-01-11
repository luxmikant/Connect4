package game

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"connect4-multiplayer/pkg/models"
)

// MockGameSessionRepository is a mock implementation of GameSessionRepository
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) Create(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetByID(ctx context.Context, id string) (*models.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) Update(ctx context.Context, session *models.GameSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetActiveGames(ctx context.Context) ([]*models.GameSession, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetGamesByPlayer(ctx context.Context, playerID string) ([]*models.GameSession, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetGameHistory(ctx context.Context, limit, offset int) ([]*models.GameSession, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetActiveSessionCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockGameSessionRepository) GetTimedOutSessions(ctx context.Context, timeout time.Duration) ([]*models.GameSession, error) {
	args := m.Called(ctx, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) BulkUpdateStatus(ctx context.Context, sessionIDs []string, status models.GameStatus) error {
	args := m.Called(ctx, sessionIDs, status)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	args := m.Called(ctx, roomCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameSession), args.Error(1)
}

// MockPlayerStatsRepository is a mock implementation of PlayerStatsRepository
type MockPlayerStatsRepository struct {
	mock.Mock
}

func (m *MockPlayerStatsRepository) Create(ctx context.Context, stats *models.PlayerStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) GetByID(ctx context.Context, id string) (*models.PlayerStats, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) GetByUsername(ctx context.Context, username string) (*models.PlayerStats, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) Update(ctx context.Context, stats *models.PlayerStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlayerStatsRepository) GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PlayerStats), args.Error(1)
}

func (m *MockPlayerStatsRepository) UpdateGameStats(ctx context.Context, username string, won bool, gameDuration int) error {
	args := m.Called(ctx, username, won, gameDuration)
	return args.Error(0)
}

// MockMoveRepository is a mock implementation of MoveRepository
type MockMoveRepository struct {
	mock.Mock
}

func (m *MockMoveRepository) Create(ctx context.Context, move *models.Move) error {
	args := m.Called(ctx, move)
	return args.Error(0)
}

func (m *MockMoveRepository) GetByID(ctx context.Context, id string) (*models.Move, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Move), args.Error(1)
}

func (m *MockMoveRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.Move, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Move), args.Error(1)
}

func (m *MockMoveRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMoveRepository) GetMoveHistory(ctx context.Context, gameID string, limit int) ([]*models.Move, error) {
	args := m.Called(ctx, gameID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Move), args.Error(1)
}

// MockGameEventRepository is a mock implementation of GameEventRepository
type MockGameEventRepository struct {
	mock.Mock
}

func (m *MockGameEventRepository) Create(ctx context.Context, event *models.GameEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockGameEventRepository) GetByID(ctx context.Context, id string) (*models.GameEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameEvent), args.Error(1)
}

func (m *MockGameEventRepository) GetByGameID(ctx context.Context, gameID string) ([]*models.GameEvent, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameEvent), args.Error(1)
}

func (m *MockGameEventRepository) GetByEventType(ctx context.Context, eventType models.EventType, limit, offset int) ([]*models.GameEvent, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameEvent), args.Error(1)
}

func (m *MockGameEventRepository) GetEventsByTimeRange(ctx context.Context, start, end string, limit, offset int) ([]*models.GameEvent, error) {
	args := m.Called(ctx, start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GameEvent), args.Error(1)
}

// Helper function to create a test service with mocks
func createTestService() (*gameService, *MockGameSessionRepository, *MockPlayerStatsRepository, *MockMoveRepository, *MockGameEventRepository) {
	gameRepo := new(MockGameSessionRepository)
	statsRepo := new(MockPlayerStatsRepository)
	moveRepo := new(MockMoveRepository)
	eventRepo := new(MockGameEventRepository)

	config := &ServiceConfig{
		SessionTimeout:    30 * time.Minute,
		DisconnectTimeout: 30 * time.Second,
		Logger:            slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	service := NewGameService(gameRepo, statsRepo, moveRepo, eventRepo, config).(*gameService)
	return service, gameRepo, statsRepo, moveRepo, eventRepo
}

func TestCreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("creates session with valid players", func(t *testing.T) {
		service, gameRepo, _, _, eventRepo := createTestService()
		gameRepo.On("Create", ctx, mock.AnythingOfType("*models.GameSession")).Return(nil).Once()
		eventRepo.On("Create", ctx, mock.AnythingOfType("*models.GameEvent")).Return(nil).Once()

		session, err := service.CreateSession(ctx, "player1", "player2")

		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "player1", session.Player1)
		assert.Equal(t, "player2", session.Player2)
		assert.Equal(t, models.StatusInProgress, session.Status)
		assert.Equal(t, models.PlayerColorRed, session.CurrentTurn)
		gameRepo.AssertExpectations(t)
	})

	t.Run("fails with empty player1", func(t *testing.T) {
		service, _, _, _, _ := createTestService()
		session, err := service.CreateSession(ctx, "", "player2")

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("fails with same players", func(t *testing.T) {
		service, _, _, _, _ := createTestService()
		session, err := service.CreateSession(ctx, "player1", "player1")

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "different usernames")
	})
}

func TestAssignPlayerColors(t *testing.T) {
	ctx := context.Background()

	t.Run("assigns red to player1 and yellow to player2", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:          "game-134",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
		}
		gameRepo.On("GetByID", ctx, "game-134").Return(session, nil).Once()

		colors, err := service.AssignPlayerColors(ctx, "game-134")

		require.NoError(t, err)
		assert.Equal(t, models.PlayerColorRed, colors["alice"])
		assert.Equal(t, models.PlayerColorYellow, colors["bob"])
	})
}

func TestGetCurrentTurn(t *testing.T) {
	ctx := context.Background()

	t.Run("returns player1 when red's turn", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:          "game-123",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
		}
		gameRepo.On("GetByID", ctx, "game-123").Return(session, nil).Once()

		player, color, err := service.GetCurrentTurn(ctx, "game-123")

		require.NoError(t, err)
		assert.Equal(t, "alice", player)
		assert.Equal(t, models.PlayerColorRed, color)
	})

	t.Run("returns player2 when yellow's turn", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:          "game-124",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorYellow,
		}
		gameRepo.On("GetByID", ctx, "game-124").Return(session, nil).Once()

		player, color, err := service.GetCurrentTurn(ctx, "game-124")

		require.NoError(t, err)
		assert.Equal(t, "bob", player)
		assert.Equal(t, models.PlayerColorYellow, color)
	})
}

func TestSwitchTurn(t *testing.T) {
	ctx := context.Background()

	t.Run("switches from red to yellow", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:          "game-125",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
		}
		gameRepo.On("GetByID", ctx, "game-125").Return(session, nil).Once()
		gameRepo.On("Update", ctx, mock.AnythingOfType("*models.GameSession")).Return(nil).Once()

		err := service.SwitchTurn(ctx, "game-125")

		require.NoError(t, err)
		assert.Equal(t, models.PlayerColorYellow, session.CurrentTurn)
	})

	t.Run("fails for inactive game", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:          "game-126",
			Status:      models.StatusCompleted,
			CurrentTurn: models.PlayerColorRed,
		}
		gameRepo.On("GetByID", ctx, "game-126").Return(session, nil).Once()

		err := service.SwitchTurn(ctx, "game-126")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})
}

func TestCompleteGame(t *testing.T) {
	ctx := context.Background()

	t.Run("completes game with winner", func(t *testing.T) {
		service, gameRepo, statsRepo, _, eventRepo := createTestService()
		session := &models.GameSession{
			ID:          "game-127",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
			StartTime:   time.Now().Add(-5 * time.Minute),
		}
		winner := models.PlayerColorRed

		gameRepo.On("GetByID", ctx, "game-127").Return(session, nil).Once()
		gameRepo.On("Update", ctx, mock.AnythingOfType("*models.GameSession")).Return(nil).Once()
		statsRepo.On("UpdateGameStats", ctx, "alice", true, mock.AnythingOfType("int")).Return(nil).Once()
		statsRepo.On("UpdateGameStats", ctx, "bob", false, mock.AnythingOfType("int")).Return(nil).Once()
		eventRepo.On("Create", ctx, mock.AnythingOfType("*models.GameEvent")).Return(nil).Once()

		err := service.CompleteGame(ctx, "game-127", &winner)

		require.NoError(t, err)
		assert.Equal(t, models.StatusCompleted, session.Status)
		assert.NotNil(t, session.Winner)
		assert.Equal(t, models.PlayerColorRed, *session.Winner)
		assert.NotNil(t, session.EndTime)
	})

	t.Run("fails for inactive game", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:     "game-128",
			Status: models.StatusCompleted,
		}
		gameRepo.On("GetByID", ctx, "game-128").Return(session, nil).Once()

		err := service.CompleteGame(ctx, "game-128", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})
}

func TestPlayerDisconnection(t *testing.T) {
	ctx := context.Background()

	t.Run("marks player as disconnected", func(t *testing.T) {
		service, gameRepo, _, _, eventRepo := createTestService()
		session := &models.GameSession{
			ID:          "game-129",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
		}
		gameRepo.On("GetByID", ctx, "game-129").Return(session, nil).Once()
		eventRepo.On("Create", ctx, mock.AnythingOfType("*models.GameEvent")).Return(nil).Once()

		err := service.MarkPlayerDisconnected(ctx, "game-129", "alice")

		require.NoError(t, err)
		assert.True(t, service.IsPlayerDisconnected("game-129", "alice"))
	})

	t.Run("marks player as reconnected", func(t *testing.T) {
		service, gameRepo, _, _, eventRepo := createTestService()
		session := &models.GameSession{
			ID:          "game-130",
			Player1:     "alice",
			Player2:     "bob",
			Status:      models.StatusInProgress,
			CurrentTurn: models.PlayerColorRed,
		}

		// First disconnect
		gameRepo.On("GetByID", ctx, "game-130").Return(session, nil).Twice()
		eventRepo.On("Create", ctx, mock.AnythingOfType("*models.GameEvent")).Return(nil).Twice()

		err := service.MarkPlayerDisconnected(ctx, "game-130", "alice")
		require.NoError(t, err)

		// Then reconnect
		err = service.MarkPlayerReconnected(ctx, "game-130", "alice")
		require.NoError(t, err)
		assert.False(t, service.IsPlayerDisconnected("game-130", "alice"))
	})

	t.Run("fails for player not in game", func(t *testing.T) {
		service, gameRepo, _, _, _ := createTestService()
		session := &models.GameSession{
			ID:      "game-131",
			Player1: "alice",
			Player2: "bob",
			Status:  models.StatusInProgress,
		}
		gameRepo.On("GetByID", ctx, "game-131").Return(session, nil).Once()

		err := service.MarkPlayerDisconnected(ctx, "game-131", "charlie")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not part of game")
	})
}

func TestGetDisconnectionTimeRemaining(t *testing.T) {
	ctx := context.Background()

	t.Run("returns remaining time for disconnected player", func(t *testing.T) {
		service, gameRepo, _, _, eventRepo := createTestService()
		session := &models.GameSession{
			ID:      "game-132",
			Player1: "alice",
			Player2: "bob",
			Status:  models.StatusInProgress,
		}
		gameRepo.On("GetByID", ctx, "game-132").Return(session, nil).Once()
		eventRepo.On("Create", ctx, mock.AnythingOfType("*models.GameEvent")).Return(nil).Once()

		err := service.MarkPlayerDisconnected(ctx, "game-132", "alice")
		require.NoError(t, err)

		remaining := service.GetDisconnectionTimeRemaining("game-132", "alice")
		assert.True(t, remaining > 0)
		assert.True(t, remaining <= 30*time.Second)
	})

	t.Run("returns zero for connected player", func(t *testing.T) {
		service, _, _, _, _ := createTestService()
		remaining := service.GetDisconnectionTimeRemaining("game-133", "bob")
		assert.Equal(t, time.Duration(0), remaining)
	})
}

func TestCacheOperations(t *testing.T) {
	service, _, _, _, _ := createTestService()

	t.Run("caches and retrieves session", func(t *testing.T) {
		session := &models.GameSession{
			ID:      "game-123",
			Player1: "alice",
			Player2: "bob",
			Status:  models.StatusInProgress,
		}

		service.CacheSession(session)
		cached, ok := service.GetCachedSession("game-123")

		assert.True(t, ok)
		assert.Equal(t, session.ID, cached.ID)
	})

	t.Run("invalidates cache", func(t *testing.T) {
		session := &models.GameSession{
			ID:      "game-456",
			Player1: "alice",
			Player2: "bob",
			Status:  models.StatusInProgress,
		}

		service.CacheSession(session)
		service.InvalidateCache("game-456")
		_, ok := service.GetCachedSession("game-456")

		assert.False(t, ok)
	})

	t.Run("returns cache stats", func(t *testing.T) {
		session := &models.GameSession{
			ID:      "game-789",
			Player1: "alice",
			Player2: "bob",
			Status:  models.StatusInProgress,
		}

		service.CacheSession(session)
		stats := service.GetCacheStats()

		assert.Contains(t, stats, "cached_sessions")
		assert.True(t, stats["cached_sessions"].(int) >= 1)
	})
}
