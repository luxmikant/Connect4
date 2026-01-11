package matchmaking

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/pkg/models"
)

// MatchmakingService defines the interface for player matchmaking
type MatchmakingService interface {
	// Queue management
	JoinQueue(ctx context.Context, username string) (*QueueEntry, error)
	LeaveQueue(ctx context.Context, username string) error
	GetQueueStatus(ctx context.Context, username string) (*QueueStatus, error)
	GetQueueLength(ctx context.Context) int

	// Matchmaking operations
	StartMatchmaking(ctx context.Context) error
	StopMatchmaking()

	// Direct bot game creation
	CreateBotGame(ctx context.Context, player string) (*models.GameSession, error)

	// Event callbacks
	SetGameCreatedCallback(callback GameCreatedCallback)
	SetBotGameCallback(callback BotGameCallback)
}

// QueueEntry represents a player in the matchmaking queue
type QueueEntry struct {
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joinedAt"`
	Timeout  time.Time `json:"timeout"`
}

// QueueStatus represents the current status of a player in the queue
type QueueStatus struct {
	InQueue       bool          `json:"inQueue"`
	Position      int           `json:"position"`
	WaitTime      time.Duration `json:"waitTime"`
	TimeRemaining time.Duration `json:"timeRemaining"`
}

// GameCreatedCallback is called when a game is created between two players
type GameCreatedCallback func(ctx context.Context, player1, player2 string, gameSession *models.GameSession) error

// BotGameCallback is called when a player is matched with a bot
type BotGameCallback func(ctx context.Context, player string, gameSession *models.GameSession) error

// matchmakingService implements MatchmakingService interface
type matchmakingService struct {
	gameService game.GameService

	// Queue management
	queue       []*QueueEntry
	queueMutex  sync.RWMutex
	playerIndex map[string]int // username -> queue position

	// Configuration
	matchTimeout  time.Duration // 10 seconds per requirement
	matchInterval time.Duration // How often to check for matches
	logger        *slog.Logger

	// Worker control
	matchWorkerCancel context.CancelFunc
	matchWorkerWg     sync.WaitGroup

	// Event callbacks
	gameCreatedCallback GameCreatedCallback
	botGameCallback     BotGameCallback
}

// ServiceConfig holds configuration for the matchmaking service
type ServiceConfig struct {
	MatchTimeout  time.Duration
	MatchInterval time.Duration
	Logger        *slog.Logger
}

// DefaultServiceConfig returns default matchmaking service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MatchTimeout:  10 * time.Second, // Requirement 1.3: 10-second timeout
		MatchInterval: 1 * time.Second,  // Check for matches every second
		Logger:        slog.Default(),
	}
}

// NewMatchmakingService creates a new MatchmakingService instance
func NewMatchmakingService(
	gameService game.GameService,
	config *ServiceConfig,
) MatchmakingService {
	if config == nil {
		config = DefaultServiceConfig()
	}

	return &matchmakingService{
		gameService:   gameService,
		queue:         make([]*QueueEntry, 0),
		playerIndex:   make(map[string]int),
		matchTimeout:  config.MatchTimeout,
		matchInterval: config.MatchInterval,
		logger:        config.Logger,
	}
}

// JoinQueue adds a player to the matchmaking queue
// Implements Requirement 1.1: add player to queue when requesting a game
func (s *matchmakingService) JoinQueue(ctx context.Context, username string) (*QueueEntry, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	// Check if player is already in queue
	if _, exists := s.playerIndex[username]; exists {
		s.logger.Info("player already in queue, ignoring duplicate join", "username", username)
		// Return the existing entry instead of error (graceful handling)
		pos := s.playerIndex[username]
		return s.queue[pos], nil
	}

	// Validate that username is unique within active sessions (Requirement 1.5)
	activeSession, err := s.gameService.GetActiveSessionByPlayer(ctx, username)
	if err == nil && activeSession != nil {
		return nil, fmt.Errorf("player %s is already in an active game", username)
	}

	// Create queue entry
	now := time.Now()
	entry := &QueueEntry{
		Username: username,
		JoinedAt: now,
		Timeout:  now.Add(s.matchTimeout),
	}

	// Add to queue
	s.queue = append(s.queue, entry)
	s.playerIndex[username] = len(s.queue) - 1

	s.logger.Info("player joined matchmaking queue",
		"username", username,
		"queueLength", len(s.queue),
		"timeout", s.matchTimeout.String(),
	)

	return entry, nil
}

// LeaveQueue removes a player from the matchmaking queue
func (s *matchmakingService) LeaveQueue(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	position, exists := s.playerIndex[username]
	if !exists {
		s.logger.Info("player not in queue, ignoring duplicate leave", "username", username)
		return nil // Graceful handling - not an error
	}

	// Remove from queue
	s.removeFromQueue(position)

	s.logger.Info("player left matchmaking queue",
		"username", username,
		"queueLength", len(s.queue),
	)

	return nil
}

// removeFromQueue removes a player at the specified position (must hold lock)
func (s *matchmakingService) removeFromQueue(position int) {
	if position < 0 || position >= len(s.queue) {
		return
	}

	username := s.queue[position].Username

	// Remove from slice
	s.queue = append(s.queue[:position], s.queue[position+1:]...)

	// Rebuild index
	delete(s.playerIndex, username)
	for i, entry := range s.queue {
		s.playerIndex[entry.Username] = i
	}
}

// GetQueueStatus returns the current status of a player in the queue
func (s *matchmakingService) GetQueueStatus(ctx context.Context, username string) (*QueueStatus, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	s.queueMutex.RLock()
	defer s.queueMutex.RUnlock()

	position, exists := s.playerIndex[username]
	if !exists {
		return &QueueStatus{
			InQueue: false,
		}, nil
	}

	entry := s.queue[position]
	now := time.Now()
	waitTime := now.Sub(entry.JoinedAt)
	timeRemaining := entry.Timeout.Sub(now)
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	return &QueueStatus{
		InQueue:       true,
		Position:      position + 1, // 1-based position
		WaitTime:      waitTime,
		TimeRemaining: timeRemaining,
	}, nil
}

// GetQueueLength returns the current number of players in the queue
func (s *matchmakingService) GetQueueLength(ctx context.Context) int {
	s.queueMutex.RLock()
	defer s.queueMutex.RUnlock()

	return len(s.queue)
}

// StartMatchmaking starts the background matchmaking worker
func (s *matchmakingService) StartMatchmaking(ctx context.Context) error {
	if s.matchWorkerCancel != nil {
		return fmt.Errorf("matchmaking is already running")
	}

	matchCtx, cancel := context.WithCancel(ctx)
	s.matchWorkerCancel = cancel

	s.matchWorkerWg.Add(1)
	go func() {
		defer s.matchWorkerWg.Done()

		ticker := time.NewTicker(s.matchInterval)
		defer ticker.Stop()

		s.logger.Info("matchmaking worker started",
			"matchTimeout", s.matchTimeout.String(),
			"matchInterval", s.matchInterval.String(),
		)

		for {
			select {
			case <-matchCtx.Done():
				s.logger.Info("matchmaking worker stopped")
				return
			case <-ticker.C:
				s.processMatchmaking(matchCtx)
			}
		}
	}()

	return nil
}

// StopMatchmaking stops the background matchmaking worker
func (s *matchmakingService) StopMatchmaking() {
	if s.matchWorkerCancel != nil {
		s.matchWorkerCancel()
		s.matchWorkerWg.Wait()
		s.matchWorkerCancel = nil
	}
}

// processMatchmaking handles the core matchmaking logic
func (s *matchmakingService) processMatchmaking(ctx context.Context) {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	now := time.Now()
	var toRemove []int

	// Process queue from oldest to newest
	for i := 0; i < len(s.queue); i++ {
		entry := s.queue[i]

		// Check if player has timed out (Requirement 1.3: 10-second timeout)
		if now.After(entry.Timeout) {
			// Start bot game
			if err := s.createBotGame(ctx, entry.Username); err != nil {
				s.logger.Error("failed to create bot game",
					"username", entry.Username,
					"error", err,
				)
			}
			toRemove = append(toRemove, i)
			continue
		}

		// Try to find a match with another player (Requirement 1.2)
		for j := i + 1; j < len(s.queue); j++ {
			otherEntry := s.queue[j]

			// Create game between the two players
			if err := s.createPlayerGame(ctx, entry.Username, otherEntry.Username); err != nil {
				s.logger.Error("failed to create player game",
					"player1", entry.Username,
					"player2", otherEntry.Username,
					"error", err,
				)
				continue
			}

			// Mark both players for removal
			toRemove = append(toRemove, i, j)
			break
		}

		// If we found a match, skip to next iteration
		if len(toRemove) > 0 && toRemove[len(toRemove)-1] == i {
			break
		}
	}

	// Remove matched/timed-out players (in reverse order to maintain indices)
	for i := len(toRemove) - 1; i >= 0; i-- {
		s.removeFromQueue(toRemove[i])
	}
}

// createPlayerGame creates a game between two players
// Implements Requirement 1.2: create game session and notify both players
func (s *matchmakingService) createPlayerGame(ctx context.Context, player1, player2 string) error {
	// Create game session (Requirement 1.4: assign colors and turn order)
	gameSession, err := s.gameService.CreateSession(ctx, player1, player2)
	if err != nil {
		return fmt.Errorf("failed to create game session: %w", err)
	}

	s.logger.Info("created player vs player game",
		"gameID", gameSession.ID,
		"player1", player1,
		"player2", player2,
	)

	// Notify via callback if set
	if s.gameCreatedCallback != nil {
		if err := s.gameCreatedCallback(ctx, player1, player2, gameSession); err != nil {
			s.logger.Warn("game created callback failed",
				"gameID", gameSession.ID,
				"error", err,
			)
		}
	}

	return nil
}

// createBotGame creates a game between a player and a bot
// Implements Requirement 1.3: start bot game after 10-second timeout
func (s *matchmakingService) createBotGame(ctx context.Context, player string) error {
	botUsername := "bot_" + generateBotID()

	// Create game session with bot
	gameSession, err := s.gameService.CreateSession(ctx, player, botUsername)
	if err != nil {
		return fmt.Errorf("failed to create bot game session: %w", err)
	}

	s.logger.Info("created player vs bot game",
		"gameID", gameSession.ID,
		"player", player,
		"bot", botUsername,
	)

	// Notify via callback if set
	if s.botGameCallback != nil {
		if err := s.botGameCallback(ctx, player, gameSession); err != nil {
			s.logger.Warn("bot game callback failed",
				"gameID", gameSession.ID,
				"error", err,
			)
		}
	}

	return nil
}

// CreateBotGame creates a game between a player and a bot (public method)
// Returns the created game session
func (s *matchmakingService) CreateBotGame(ctx context.Context, player string) (*models.GameSession, error) {
	botUsername := "bot_" + generateBotID()

	// Create game session with bot
	gameSession, err := s.gameService.CreateSession(ctx, player, botUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot game session: %w", err)
	}

	s.logger.Info("created player vs bot game (direct)",
		"gameID", gameSession.ID,
		"player", player,
		"bot", botUsername,
	)

	return gameSession, nil
}

// SetGameCreatedCallback sets the callback for when a player vs player game is created
func (s *matchmakingService) SetGameCreatedCallback(callback GameCreatedCallback) {
	s.gameCreatedCallback = callback
}

// SetBotGameCallback sets the callback for when a player vs bot game is created
func (s *matchmakingService) SetBotGameCallback(callback BotGameCallback) {
	s.botGameCallback = callback
}

// generateBotID generates a unique ID for bot players
func generateBotID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}
