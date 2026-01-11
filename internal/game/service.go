package game

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

// AnalyticsProducer interface for sending analytics events to Kafka
type AnalyticsProducer interface {
	SendGameStarted(ctx context.Context, gameID, player1, player2 string) error
	SendGameCompleted(ctx context.Context, gameID, winner, loser string, duration time.Duration) error
	SendMoveMade(ctx context.Context, gameID, playerID string, column, row int, moveNumber int) error
	SendPlayerJoined(ctx context.Context, gameID, playerID string) error
	SendPlayerDisconnected(ctx context.Context, gameID, playerID string) error
	SendPlayerReconnected(ctx context.Context, gameID, playerID string) error
}

// GameService defines the interface for game session management
type GameService interface {
	// Session lifecycle management
	CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error)
	GetSession(ctx context.Context, gameID string) (*models.GameSession, error)
	EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error
	
	// Custom room management
	CreateCustomRoom(ctx context.Context, creator string) (*models.GameSession, string, error)
	JoinCustomRoom(ctx context.Context, roomCode, username string) (*models.GameSession, error)
	GetSessionByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error)
	
	// Turn management
	GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error)
	SwitchTurn(ctx context.Context, gameID string) error
	
	// Player color assignment
	AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error)
	
	// Game completion and statistics
	CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error
	
	// Active session management
	GetActiveSessions(ctx context.Context) ([]*models.GameSession, error)
	GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error)
	GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error)
	GetActiveSessionCount(ctx context.Context) (int64, error)
	
	// Session cleanup and timeout handling
	CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error)
	MarkSessionAbandoned(ctx context.Context, gameID string) error
	StartCleanupWorker(ctx context.Context, interval time.Duration)
	StopCleanupWorker()
	
	// Player disconnection handling (Requirement 4)
	MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error
	MarkPlayerReconnected(ctx context.Context, gameID string, username string) error
	GetDisconnectedPlayers(gameID string) map[string]time.Time
	HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error
	
	// In-memory cache operations
	CacheSession(session *models.GameSession)
	GetCachedSession(gameID string) (*models.GameSession, bool)
	InvalidateCache(gameID string)
	CleanupCache(maxAge time.Duration) int
	GetCacheStats() map[string]interface{}
}

// gameService implements GameService interface
type gameService struct {
	gameRepo  repositories.GameSessionRepository
	statsRepo repositories.PlayerStatsRepository
	moveRepo  repositories.MoveRepository
	eventRepo repositories.GameEventRepository
	
	// Analytics producer for Kafka events (Requirement 9, 10)
	analyticsProducer AnalyticsProducer
	
	// In-memory cache for active sessions
	sessionCache map[string]*cachedSession
	cacheMutex   sync.RWMutex
	
	// Player disconnection tracking (Requirement 4)
	disconnectedPlayers map[string]map[string]time.Time // gameID -> username -> disconnectTime
	disconnectMutex     sync.RWMutex
	
	// Cleanup worker control
	cleanupCancel context.CancelFunc
	cleanupWg     sync.WaitGroup
	
	// Configuration
	sessionTimeout       time.Duration
	disconnectTimeout    time.Duration // 30 seconds per Requirement 4
	logger               *slog.Logger
}

// cachedSession wraps a game session with cache metadata
type cachedSession struct {
	Session    *models.GameSession
	CachedAt   time.Time
	LastAccess time.Time
}

// ServiceConfig holds configuration for the game service
type ServiceConfig struct {
	SessionTimeout    time.Duration
	DisconnectTimeout time.Duration
	Logger            *slog.Logger
	AnalyticsProducer AnalyticsProducer // Optional: Kafka producer for analytics
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		SessionTimeout:    30 * time.Minute,
		DisconnectTimeout: 30 * time.Second, // Requirement 4: 30 second timeout
		Logger:            slog.Default(),
		AnalyticsProducer: nil, // Optional, can be set later
	}
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repositories.GameSessionRepository,
	statsRepo repositories.PlayerStatsRepository,
	moveRepo repositories.MoveRepository,
	eventRepo repositories.GameEventRepository,
	config *ServiceConfig,
) GameService {
	if config == nil {
		config = DefaultServiceConfig()
	}
	
	return &gameService{
		gameRepo:            gameRepo,
		statsRepo:           statsRepo,
		moveRepo:            moveRepo,
		eventRepo:           eventRepo,
		analyticsProducer:   config.AnalyticsProducer,
		sessionCache:        make(map[string]*cachedSession),
		disconnectedPlayers: make(map[string]map[string]time.Time),
		sessionTimeout:      config.SessionTimeout,
		disconnectTimeout:   config.DisconnectTimeout,
		logger:              config.Logger,
	}
}

// CreateSession creates a new game session with player color assignment
func (s *gameService) CreateSession(ctx context.Context, player1, player2 string) (*models.GameSession, error) {
	if player1 == "" || player2 == "" {
		return nil, fmt.Errorf("player usernames cannot be empty")
	}
	
	if player1 == player2 {
		return nil, fmt.Errorf("players must have different usernames")
	}
	
	// Create new game session
	session := &models.GameSession{
		Player1:     player1,
		Player2:     player2,
		Status:      models.StatusInProgress,
		CurrentTurn: models.PlayerColorRed, // Player1 (red) always starts
		Board:       models.NewBoard(),
		StartTime:   time.Now(),
	}
	
	// Persist to database
	if err := s.gameRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}
	
	// Cache the session for fast access
	s.CacheSession(session)
	
	// Log session creation
	s.logger.Info("game session created",
		"gameID", session.ID,
		"player1", player1,
		"player2", player2,
	)
	
	// Create game started event in database
	event := models.NewGameStartedEvent(session.ID, player1, player2)
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Warn("failed to create game started event",
			"gameID", session.ID,
			"error", err,
		)
	}
	
	// Send analytics event to Kafka (Requirement 9.1)
	if s.analyticsProducer != nil {
		go func() {
			if err := s.analyticsProducer.SendGameStarted(context.Background(), session.ID, player1, player2); err != nil {
				s.logger.Warn("failed to send game started analytics event",
					"gameID", session.ID,
					"error", err,
				)
			}
		}()
	}
	
	return session, nil
}

// generateRoomCode generates a unique 8-character alphanumeric room code
func generateRoomCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8
	
	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = charset[num.Int64()]
	}
	
	return string(code), nil
}

// CreateCustomRoom creates a new custom game room with a unique room code
func (s *gameService) CreateCustomRoom(ctx context.Context, creator string) (*models.GameSession, string, error) {
	if creator == "" {
		return nil, "", fmt.Errorf("creator username cannot be empty")
	}
	
	// Generate unique room code (retry up to 5 times if collision)
	var roomCode string
	var err error
	for i := 0; i < 5; i++ {
		roomCode, err = generateRoomCode()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate room code: %w", err)
		}
		
		// Check if room code already exists
		existing, _ := s.GetSessionByRoomCode(ctx, roomCode)
		if existing == nil {
			break // Code is unique
		}
		
		if i == 4 {
			return nil, "", fmt.Errorf("failed to generate unique room code after 5 attempts")
		}
	}
	
	// Create new game session with custom room fields
	// Use temporary placeholder for player2 until opponent joins
	session := &models.GameSession{
		Player1:     creator,
		Player2:     "waiting", // Placeholder until opponent joins
		Status:      models.StatusWaiting,
		CurrentTurn: models.PlayerColorRed,
		Board:       models.NewBoard(),
		StartTime:   time.Now(),
		RoomCode:    &roomCode,
		IsCustom:    true,
		CreatedBy:   &creator,
	}
	
	// Persist to database
	if err := s.gameRepo.Create(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create custom room: %w", err)
	}
	
	// Cache the session for fast access
	s.CacheSession(session)
	
	// Log room creation
	s.logger.Info("custom room created",
		"gameID", session.ID,
		"roomCode", roomCode,
		"creator", creator,
	)
	
	return session, roomCode, nil
}

// JoinCustomRoom allows a player to join an existing custom room by code
func (s *gameService) JoinCustomRoom(ctx context.Context, roomCode, username string) (*models.GameSession, error) {
	if roomCode == "" {
		return nil, fmt.Errorf("room code cannot be empty")
	}
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	
	// Normalize room code to uppercase
	roomCode = strings.ToUpper(strings.TrimSpace(roomCode))
	
	// Find session by room code
	session, err := s.GetSessionByRoomCode(ctx, roomCode)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}
	
	if session == nil {
		return nil, fmt.Errorf("room with code %s does not exist", roomCode)
	}
	
	// Validate room is still waiting for players
	if session.Status != models.StatusWaiting {
		return nil, fmt.Errorf("room is no longer available (status: %s)", session.Status)
	}
	
	// Check if user is trying to join their own room
	if session.Player1 == username {
		return nil, fmt.Errorf("cannot join your own room")
	}
	
	// Check if room already has a second player
	if session.Player2 != "waiting" {
		return nil, fmt.Errorf("room is already full")
	}
	
	// Add player as Player2 and start the game
	session.Player2 = username
	session.Status = models.StatusInProgress
	session.StartTime = time.Now()
	
	// Persist changes
	if err := s.gameRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to join custom room: %w", err)
	}
	
	// Update cache
	s.CacheSession(session)
	
	// Log player join
	s.logger.Info("player joined custom room",
		"gameID", session.ID,
		"roomCode", roomCode,
		"player1", session.Player1,
		"player2", username,
	)
	
	// Create game started event
	event := models.NewGameStartedEvent(session.ID, session.Player1, username)
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Warn("failed to create game started event",
			"gameID", session.ID,
			"error", err,
		)
	}
	
	// Send analytics event to Kafka
	if s.analyticsProducer != nil {
		go func() {
			if err := s.analyticsProducer.SendGameStarted(context.Background(), session.ID, session.Player1, username); err != nil {
				s.logger.Warn("failed to send game started analytics event",
					"gameID", session.ID,
					"error", err,
				)
			}
		}()
	}
	
	return session, nil
}

// GetSessionByRoomCode retrieves a game session by room code
func (s *gameService) GetSessionByRoomCode(ctx context.Context, roomCode string) (*models.GameSession, error) {
	if roomCode == "" {
		return nil, fmt.Errorf("room code cannot be empty")
	}
	
	// Normalize room code
	roomCode = strings.ToUpper(strings.TrimSpace(roomCode))
	
	// Query database for session with this room code
	// Note: This requires adding a method to the repository
	sessions, err := s.gameRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	
	for _, session := range sessions {
		if session.RoomCode != nil && *session.RoomCode == roomCode {
			return session, nil
		}
	}
	
	return nil, nil
}
}


// GetSession retrieves a game session, checking cache first
func (s *gameService) GetSession(ctx context.Context, gameID string) (*models.GameSession, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}
	
	// Check cache first
	if cached, ok := s.GetCachedSession(gameID); ok {
		return cached, nil
	}
	
	// Fetch from database
	session, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	
	// Cache active sessions
	if session.IsActive() {
		s.CacheSession(session)
	}
	
	return session, nil
}

// EndSession ends a game session with the specified outcome
func (s *gameService) EndSession(ctx context.Context, gameID string, winner *models.PlayerColor, reason string) error {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return fmt.Errorf("game session is not active")
	}
	
	// Update session status
	now := time.Now()
	session.Status = models.StatusCompleted
	session.Winner = winner
	session.EndTime = &now
	
	// Persist changes
	if err := s.gameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to end game session: %w", err)
	}
	
	// Invalidate cache
	s.InvalidateCache(gameID)
	
	s.logger.Info("game session ended",
		"gameID", gameID,
		"winner", winner,
		"reason", reason,
	)
	
	return nil
}

// GetCurrentTurn returns the current player's username and color
func (s *gameService) GetCurrentTurn(ctx context.Context, gameID string) (string, models.PlayerColor, error) {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return "", "", err
	}
	
	currentPlayer := session.GetCurrentPlayer()
	return currentPlayer, session.CurrentTurn, nil
}

// SwitchTurn switches the turn to the other player
func (s *gameService) SwitchTurn(ctx context.Context, gameID string) error {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return fmt.Errorf("cannot switch turn: game is not active")
	}
	
	// Switch turn
	if session.CurrentTurn == models.PlayerColorRed {
		session.CurrentTurn = models.PlayerColorYellow
	} else {
		session.CurrentTurn = models.PlayerColorRed
	}
	
	// Update in database
	if err := s.gameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to switch turn: %w", err)
	}
	
	// Update cache
	s.CacheSession(session)
	
	return nil
}

// AssignPlayerColors returns the color assignment for both players
func (s *gameService) AssignPlayerColors(ctx context.Context, gameID string) (map[string]models.PlayerColor, error) {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return nil, err
	}
	
	colors := map[string]models.PlayerColor{
		session.Player1: models.PlayerColorRed,    // Player1 is always red
		session.Player2: models.PlayerColorYellow, // Player2 is always yellow
	}
	
	return colors, nil
}

// CompleteGame completes a game and updates player statistics
func (s *gameService) CompleteGame(ctx context.Context, gameID string, winner *models.PlayerColor) error {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return fmt.Errorf("game is not active")
	}
	
	// Calculate game duration
	now := time.Now()
	gameDuration := int(now.Sub(session.StartTime).Seconds())
	
	// Update session
	session.Status = models.StatusCompleted
	session.Winner = winner
	session.EndTime = &now
	
	// Persist session changes
	if err := s.gameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to complete game: %w", err)
	}
	
	// Update player statistics
	if err := s.updatePlayerStats(ctx, session, winner, gameDuration); err != nil {
		s.logger.Warn("failed to update player stats",
			"gameID", gameID,
			"error", err,
		)
	}
	
	// Create game completed event
	winnerUsername := ""
	loserUsername := ""
	if winner != nil {
		if *winner == models.PlayerColorRed {
			winnerUsername = session.Player1
			loserUsername = session.Player2
		} else {
			winnerUsername = session.Player2
			loserUsername = session.Player1
		}
	}
	
	event := models.NewGameCompletedEvent(gameID, winnerUsername, loserUsername, gameDuration)
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Warn("failed to create game completed event",
			"gameID", gameID,
			"error", err,
		)
	}
	
	// Send analytics event to Kafka (Requirement 9.3)
	if s.analyticsProducer != nil {
		go func() {
			duration := time.Duration(gameDuration) * time.Second
			if err := s.analyticsProducer.SendGameCompleted(context.Background(), gameID, winnerUsername, loserUsername, duration); err != nil {
				s.logger.Warn("failed to send game completed analytics event",
					"gameID", gameID,
					"error", err,
				)
			}
		}()
	}
	
	// Invalidate cache
	s.InvalidateCache(gameID)
	
	s.logger.Info("game completed",
		"gameID", gameID,
		"winner", winnerUsername,
		"duration", gameDuration,
	)
	
	return nil
}

// updatePlayerStats updates statistics for both players after a game
func (s *gameService) updatePlayerStats(ctx context.Context, session *models.GameSession, winner *models.PlayerColor, gameDuration int) error {
	// Determine winners and losers
	player1Won := winner != nil && *winner == models.PlayerColorRed
	player2Won := winner != nil && *winner == models.PlayerColorYellow
	
	// Update player1 stats
	if err := s.statsRepo.UpdateGameStats(ctx, session.Player1, player1Won, gameDuration); err != nil {
		return fmt.Errorf("failed to update player1 stats: %w", err)
	}
	
	// Update player2 stats
	if err := s.statsRepo.UpdateGameStats(ctx, session.Player2, player2Won, gameDuration); err != nil {
		return fmt.Errorf("failed to update player2 stats: %w", err)
	}
	
	return nil
}


// GetActiveSessions retrieves all active game sessions
func (s *gameService) GetActiveSessions(ctx context.Context) ([]*models.GameSession, error) {
	return s.gameRepo.GetActiveGames(ctx)
}

// GetSessionsByPlayer retrieves all sessions for a specific player
func (s *gameService) GetSessionsByPlayer(ctx context.Context, username string) ([]*models.GameSession, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	return s.gameRepo.GetGamesByPlayer(ctx, username)
}

// CleanupTimedOutSessions marks sessions as abandoned if they've been inactive too long
func (s *gameService) CleanupTimedOutSessions(ctx context.Context, timeout time.Duration) (int, error) {
	// Get all active sessions
	sessions, err := s.gameRepo.GetActiveGames(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get active sessions: %w", err)
	}
	
	cleanedCount := 0
	cutoffTime := time.Now().Add(-timeout)
	
	for _, session := range sessions {
		// Check if session has timed out based on last activity
		lastActivity := session.UpdatedAt
		if lastActivity.Before(cutoffTime) {
			if err := s.MarkSessionAbandoned(ctx, session.ID); err != nil {
				s.logger.Warn("failed to mark session as abandoned",
					"gameID", session.ID,
					"error", err,
				)
				continue
			}
			cleanedCount++
		}
	}
	
	if cleanedCount > 0 {
		s.logger.Info("cleaned up timed out sessions",
			"count", cleanedCount,
			"timeout", timeout.String(),
		)
	}
	
	return cleanedCount, nil
}

// MarkSessionAbandoned marks a session as abandoned
func (s *gameService) MarkSessionAbandoned(ctx context.Context, gameID string) error {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return nil // Already not active
	}
	
	now := time.Now()
	session.Status = models.StatusAbandoned
	session.EndTime = &now
	
	if err := s.gameRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to mark session as abandoned: %w", err)
	}
	
	// Invalidate cache
	s.InvalidateCache(gameID)
	
	s.logger.Info("session marked as abandoned",
		"gameID", gameID,
	)
	
	return nil
}

// CacheSession adds or updates a session in the cache
func (s *gameService) CacheSession(session *models.GameSession) {
	if session == nil {
		return
	}
	
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.sessionCache[session.ID] = &cachedSession{
		Session:    session,
		CachedAt:   time.Now(),
		LastAccess: time.Now(),
	}
}

// GetCachedSession retrieves a session from cache
func (s *gameService) GetCachedSession(gameID string) (*models.GameSession, bool) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	cached, ok := s.sessionCache[gameID]
	if !ok {
		return nil, false
	}
	
	// Update last access time (requires write lock, so we skip for now)
	return cached.Session, true
}

// InvalidateCache removes a session from the cache
func (s *gameService) InvalidateCache(gameID string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	delete(s.sessionCache, gameID)
}

// CleanupCache removes stale entries from the cache
func (s *gameService) CleanupCache(maxAge time.Duration) int {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	
	for id, cached := range s.sessionCache {
		if cached.LastAccess.Before(cutoff) {
			delete(s.sessionCache, id)
			removed++
		}
	}
	
	return removed
}

// GetCacheStats returns statistics about the session cache
func (s *gameService) GetCacheStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	return map[string]interface{}{
		"cached_sessions": len(s.sessionCache),
	}
}

// GetActiveSessionByPlayer retrieves an active session for a specific player
// Uses optimized PostgreSQL query with index on status and player columns
func (s *gameService) GetActiveSessionByPlayer(ctx context.Context, username string) (*models.GameSession, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	return s.gameRepo.GetActiveSessionByPlayer(ctx, username)
}

// GetActiveSessionCount returns the count of active sessions
// Uses optimized count query with index on status
func (s *gameService) GetActiveSessionCount(ctx context.Context) (int64, error) {
	return s.gameRepo.GetActiveSessionCount(ctx)
}

// StartCleanupWorker starts a background goroutine that periodically cleans up
// timed-out sessions and handles player disconnection timeouts
func (s *gameService) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	cleanupCtx, cancel := context.WithCancel(ctx)
	s.cleanupCancel = cancel
	
	s.cleanupWg.Add(1)
	go func() {
		defer s.cleanupWg.Done()
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		s.logger.Info("cleanup worker started",
			"interval", interval.String(),
			"sessionTimeout", s.sessionTimeout.String(),
			"disconnectTimeout", s.disconnectTimeout.String(),
		)
		
		for {
			select {
			case <-cleanupCtx.Done():
				s.logger.Info("cleanup worker stopped")
				return
			case <-ticker.C:
				// Clean up timed-out sessions
				if _, err := s.CleanupTimedOutSessions(cleanupCtx, s.sessionTimeout); err != nil {
					s.logger.Error("failed to cleanup timed out sessions", "error", err)
				}
				
				// Handle player disconnection timeouts
				s.handleDisconnectionTimeouts(cleanupCtx)
				
				// Clean up stale cache entries
				removed := s.CleanupCache(s.sessionTimeout)
				if removed > 0 {
					s.logger.Debug("cleaned up stale cache entries", "count", removed)
				}
				
				// Clean up old disconnection tracking entries
				s.cleanupDisconnectionTracking()
			}
		}
	}()
}

// StopCleanupWorker stops the background cleanup worker
func (s *gameService) StopCleanupWorker() {
	if s.cleanupCancel != nil {
		s.cleanupCancel()
		s.cleanupWg.Wait()
	}
}

// MarkPlayerDisconnected marks a player as disconnected from a game session
// Implements Requirement 4: maintain session state for 30 seconds after disconnect
func (s *gameService) MarkPlayerDisconnected(ctx context.Context, gameID string, username string) error {
	if gameID == "" || username == "" {
		return fmt.Errorf("gameID and username cannot be empty")
	}
	
	// Verify the session exists and is active
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return fmt.Errorf("game session is not active")
	}
	
	// Verify the player is part of this game
	if session.Player1 != username && session.Player2 != username {
		return fmt.Errorf("player %s is not part of game %s", username, gameID)
	}
	
	// Track disconnection time
	s.disconnectMutex.Lock()
	if s.disconnectedPlayers[gameID] == nil {
		s.disconnectedPlayers[gameID] = make(map[string]time.Time)
	}
	s.disconnectedPlayers[gameID][username] = time.Now()
	s.disconnectMutex.Unlock()
	
	// Create player left event in database
	event := &models.GameEvent{
		EventType: models.EventPlayerLeft,
		GameID:    gameID,
		PlayerID:  username,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"reason": "disconnected",
		},
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Warn("failed to create player left event",
			"gameID", gameID,
			"player", username,
			"error", err,
		)
	}
	
	// Send analytics event to Kafka (Requirement 9.4)
	if s.analyticsProducer != nil {
		go func() {
			if err := s.analyticsProducer.SendPlayerDisconnected(context.Background(), gameID, username); err != nil {
				s.logger.Warn("failed to send player disconnected analytics event",
					"gameID", gameID,
					"player", username,
					"error", err,
				)
			}
		}()
	}
	
	s.logger.Info("player disconnected",
		"gameID", gameID,
		"player", username,
		"timeout", s.disconnectTimeout.String(),
	)
	
	return nil
}

// MarkPlayerReconnected marks a player as reconnected to a game session
// Implements Requirement 4: restore connection within 30 seconds
func (s *gameService) MarkPlayerReconnected(ctx context.Context, gameID string, username string) error {
	if gameID == "" || username == "" {
		return fmt.Errorf("gameID and username cannot be empty")
	}
	
	// Verify the session exists and is active
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return fmt.Errorf("game session is not active")
	}
	
	// Remove from disconnected tracking
	s.disconnectMutex.Lock()
	if s.disconnectedPlayers[gameID] != nil {
		delete(s.disconnectedPlayers[gameID], username)
		if len(s.disconnectedPlayers[gameID]) == 0 {
			delete(s.disconnectedPlayers, gameID)
		}
	}
	s.disconnectMutex.Unlock()
	
	// Create player reconnected event in database
	event := &models.GameEvent{
		EventType: models.EventPlayerReconnected,
		GameID:    gameID,
		PlayerID:  username,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{},
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Warn("failed to create player reconnected event",
			"gameID", gameID,
			"player", username,
			"error", err,
		)
	}
	
	// Send analytics event to Kafka (Requirement 9.4)
	if s.analyticsProducer != nil {
		go func() {
			if err := s.analyticsProducer.SendPlayerReconnected(context.Background(), gameID, username); err != nil {
				s.logger.Warn("failed to send player reconnected analytics event",
					"gameID", gameID,
					"player", username,
					"error", err,
				)
			}
		}()
	}
	
	s.logger.Info("player reconnected",
		"gameID", gameID,
		"player", username,
	)
	
	return nil
}

// GetDisconnectedPlayers returns the disconnection times for players in a game
func (s *gameService) GetDisconnectedPlayers(gameID string) map[string]time.Time {
	s.disconnectMutex.RLock()
	defer s.disconnectMutex.RUnlock()
	
	if s.disconnectedPlayers[gameID] == nil {
		return nil
	}
	
	// Return a copy to avoid race conditions
	result := make(map[string]time.Time)
	for username, disconnectTime := range s.disconnectedPlayers[gameID] {
		result[username] = disconnectTime
	}
	return result
}

// HandleDisconnectionTimeout handles the timeout for a disconnected player
// Implements Requirement 4.3: forfeit game after 30 seconds
func (s *gameService) HandleDisconnectionTimeout(ctx context.Context, gameID string, username string) error {
	session, err := s.GetSession(ctx, gameID)
	if err != nil {
		return err
	}
	
	if !session.IsActive() {
		return nil // Game already ended
	}
	
	// Determine the winner (the opponent who stayed connected)
	var winner models.PlayerColor
	if session.Player1 == username {
		winner = models.PlayerColorYellow // Player2 wins
	} else {
		winner = models.PlayerColorRed // Player1 wins
	}
	
	// Complete the game with the connected player as winner
	if err := s.CompleteGame(ctx, gameID, &winner); err != nil {
		return fmt.Errorf("failed to complete game after disconnection timeout: %w", err)
	}
	
	// Remove from disconnection tracking
	s.disconnectMutex.Lock()
	if s.disconnectedPlayers[gameID] != nil {
		delete(s.disconnectedPlayers[gameID], username)
		if len(s.disconnectedPlayers[gameID]) == 0 {
			delete(s.disconnectedPlayers, gameID)
		}
	}
	s.disconnectMutex.Unlock()
	
	s.logger.Info("game forfeited due to disconnection timeout",
		"gameID", gameID,
		"disconnectedPlayer", username,
		"winner", winner,
	)
	
	return nil
}

// handleDisconnectionTimeouts checks all disconnected players and handles timeouts
func (s *gameService) handleDisconnectionTimeouts(ctx context.Context) {
	s.disconnectMutex.RLock()
	// Create a copy of the map to avoid holding the lock during processing
	toProcess := make(map[string]map[string]time.Time)
	for gameID, players := range s.disconnectedPlayers {
		toProcess[gameID] = make(map[string]time.Time)
		for username, disconnectTime := range players {
			toProcess[gameID][username] = disconnectTime
		}
	}
	s.disconnectMutex.RUnlock()
	
	now := time.Now()
	for gameID, players := range toProcess {
		for username, disconnectTime := range players {
			if now.Sub(disconnectTime) >= s.disconnectTimeout {
				if err := s.HandleDisconnectionTimeout(ctx, gameID, username); err != nil {
					s.logger.Error("failed to handle disconnection timeout",
						"gameID", gameID,
						"player", username,
						"error", err,
					)
				}
			}
		}
	}
}

// cleanupDisconnectionTracking removes entries for games that are no longer active
func (s *gameService) cleanupDisconnectionTracking() {
	s.disconnectMutex.Lock()
	defer s.disconnectMutex.Unlock()
	
	for gameID := range s.disconnectedPlayers {
		// Check if game is still active (use cached session to avoid DB call)
		if cached, ok := s.sessionCache[gameID]; ok {
			if !cached.Session.IsActive() {
				delete(s.disconnectedPlayers, gameID)
			}
		}
	}
}

// IsPlayerDisconnected checks if a player is currently marked as disconnected
func (s *gameService) IsPlayerDisconnected(gameID string, username string) bool {
	s.disconnectMutex.RLock()
	defer s.disconnectMutex.RUnlock()
	
	if s.disconnectedPlayers[gameID] == nil {
		return false
	}
	_, exists := s.disconnectedPlayers[gameID][username]
	return exists
}

// GetDisconnectionTimeRemaining returns the time remaining before a player's disconnection timeout
func (s *gameService) GetDisconnectionTimeRemaining(gameID string, username string) time.Duration {
	s.disconnectMutex.RLock()
	defer s.disconnectMutex.RUnlock()
	
	if s.disconnectedPlayers[gameID] == nil {
		return 0
	}
	
	disconnectTime, exists := s.disconnectedPlayers[gameID][username]
	if !exists {
		return 0
	}
	
	elapsed := time.Since(disconnectTime)
	remaining := s.disconnectTimeout - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}
