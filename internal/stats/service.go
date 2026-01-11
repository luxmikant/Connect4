package stats

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"connect4-multiplayer/internal/database/repositories"
	"connect4-multiplayer/pkg/models"
)

// PlayerStatsService defines the interface for player statistics management
type PlayerStatsService interface {
	// Statistics retrieval
	GetPlayerStats(ctx context.Context, username string) (*models.PlayerStats, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error)
	GetTopPlayers(ctx context.Context) ([]*models.PlayerStats, error)

	// Statistics updates
	RecordGameResult(ctx context.Context, username string, won bool, gameDuration int) error
	RecordGameCompletion(ctx context.Context, player1, player2 string, winner *models.PlayerColor, gameDuration int) error

	// Player management
	CreatePlayerStats(ctx context.Context, username string) (*models.PlayerStats, error)
	GetOrCreatePlayerStats(ctx context.Context, username string) (*models.PlayerStats, error)

	// Real-time leaderboard updates
	SubscribeToLeaderboardUpdates(callback LeaderboardUpdateCallback) string
	UnsubscribeFromLeaderboardUpdates(subscriptionID string)
	NotifyLeaderboardUpdate(ctx context.Context)

	// Cache management
	InvalidateCache(username string)
	RefreshLeaderboardCache(ctx context.Context) error
	GetCacheStats() map[string]interface{}
}

// LeaderboardUpdateCallback is called when the leaderboard is updated
type LeaderboardUpdateCallback func(leaderboard []*models.PlayerStats)

// playerStatsService implements PlayerStatsService interface
type playerStatsService struct {
	statsRepo repositories.PlayerStatsRepository
	logger    *slog.Logger

	// In-memory cache for player stats
	statsCache     map[string]*cachedStats
	statsCacheMu   sync.RWMutex
	statsCacheTTL  time.Duration

	// Leaderboard cache
	leaderboardCache    []*models.PlayerStats
	leaderboardCacheMu  sync.RWMutex
	leaderboardCacheTTL time.Duration
	leaderboardCachedAt time.Time

	// Real-time update subscriptions
	subscriptions   map[string]LeaderboardUpdateCallback
	subscriptionsMu sync.RWMutex
	nextSubID       int
}

// cachedStats wraps player stats with cache metadata
type cachedStats struct {
	Stats    *models.PlayerStats
	CachedAt time.Time
}

// ServiceConfig holds configuration for the stats service
type ServiceConfig struct {
	StatsCacheTTL       time.Duration
	LeaderboardCacheTTL time.Duration
	Logger              *slog.Logger
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		StatsCacheTTL:       5 * time.Minute,
		LeaderboardCacheTTL: 30 * time.Second, // Short TTL for real-time updates
		Logger:              slog.Default(),
	}
}

// NewPlayerStatsService creates a new PlayerStatsService instance
func NewPlayerStatsService(
	statsRepo repositories.PlayerStatsRepository,
	config *ServiceConfig,
) PlayerStatsService {
	if config == nil {
		config = DefaultServiceConfig()
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &playerStatsService{
		statsRepo:           statsRepo,
		logger:              logger,
		statsCache:          make(map[string]*cachedStats),
		statsCacheTTL:       config.StatsCacheTTL,
		leaderboardCacheTTL: config.LeaderboardCacheTTL,
		subscriptions:       make(map[string]LeaderboardUpdateCallback),
	}
}


// GetPlayerStats retrieves statistics for a specific player
func (s *playerStatsService) GetPlayerStats(ctx context.Context, username string) (*models.PlayerStats, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	// Check cache first
	if cached := s.getCachedStats(username); cached != nil {
		return cached, nil
	}

	// Fetch from database
	stats, err := s.statsRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheStats(stats)

	return stats, nil
}

// GetLeaderboard retrieves the top players sorted by wins
func (s *playerStatsService) GetLeaderboard(ctx context.Context, limit int) ([]*models.PlayerStats, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Check if we have a valid cached leaderboard
	s.leaderboardCacheMu.RLock()
	if s.leaderboardCache != nil && time.Since(s.leaderboardCachedAt) < s.leaderboardCacheTTL {
		// Return cached data (up to limit)
		result := s.leaderboardCache
		if len(result) > limit {
			result = result[:limit]
		}
		s.leaderboardCacheMu.RUnlock()
		return result, nil
	}
	s.leaderboardCacheMu.RUnlock()

	// Fetch from database
	leaderboard, err := s.statsRepo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Update cache
	s.leaderboardCacheMu.Lock()
	s.leaderboardCache = leaderboard
	s.leaderboardCachedAt = time.Now()
	s.leaderboardCacheMu.Unlock()

	return leaderboard, nil
}

// GetTopPlayers retrieves the top 10 players (convenience method)
func (s *playerStatsService) GetTopPlayers(ctx context.Context) ([]*models.PlayerStats, error) {
	return s.GetLeaderboard(ctx, 10)
}

// RecordGameResult records a game result for a single player
func (s *playerStatsService) RecordGameResult(ctx context.Context, username string, won bool, gameDuration int) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Update stats in database
	if err := s.statsRepo.UpdateGameStats(ctx, username, won, gameDuration); err != nil {
		return fmt.Errorf("failed to record game result: %w", err)
	}

	// Invalidate cache for this player
	s.InvalidateCache(username)

	// Notify subscribers of leaderboard update
	s.NotifyLeaderboardUpdate(ctx)

	s.logger.Info("game result recorded",
		"username", username,
		"won", won,
		"duration", gameDuration,
	)

	return nil
}

// RecordGameCompletion records the result of a completed game for both players
func (s *playerStatsService) RecordGameCompletion(ctx context.Context, player1, player2 string, winner *models.PlayerColor, gameDuration int) error {
	if player1 == "" || player2 == "" {
		return fmt.Errorf("player usernames cannot be empty")
	}

	// Determine who won
	player1Won := winner != nil && *winner == models.PlayerColorRed
	player2Won := winner != nil && *winner == models.PlayerColorYellow

	// Update player1 stats
	if err := s.statsRepo.UpdateGameStats(ctx, player1, player1Won, gameDuration); err != nil {
		return fmt.Errorf("failed to update player1 stats: %w", err)
	}

	// Update player2 stats
	if err := s.statsRepo.UpdateGameStats(ctx, player2, player2Won, gameDuration); err != nil {
		return fmt.Errorf("failed to update player2 stats: %w", err)
	}

	// Invalidate cache for both players
	s.InvalidateCache(player1)
	s.InvalidateCache(player2)

	// Notify subscribers of leaderboard update
	s.NotifyLeaderboardUpdate(ctx)

	s.logger.Info("game completion recorded",
		"player1", player1,
		"player2", player2,
		"winner", winner,
		"duration", gameDuration,
	)

	return nil
}

// CreatePlayerStats creates a new player stats record
func (s *playerStatsService) CreatePlayerStats(ctx context.Context, username string) (*models.PlayerStats, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	stats := &models.PlayerStats{
		Username:    username,
		GamesPlayed: 0,
		GamesWon:    0,
		WinRate:     0.0,
		AvgGameTime: 0,
		LastPlayed:  time.Now(),
	}

	if err := s.statsRepo.Create(ctx, stats); err != nil {
		return nil, fmt.Errorf("failed to create player stats: %w", err)
	}

	// Cache the new stats
	s.cacheStats(stats)

	s.logger.Info("player stats created", "username", username)

	return stats, nil
}

// GetOrCreatePlayerStats retrieves existing stats or creates new ones
func (s *playerStatsService) GetOrCreatePlayerStats(ctx context.Context, username string) (*models.PlayerStats, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	// Try to get existing stats
	stats, err := s.GetPlayerStats(ctx, username)
	if err == nil {
		return stats, nil
	}

	// If not found, create new stats
	if err == models.ErrPlayerNotFound {
		return s.CreatePlayerStats(ctx, username)
	}

	return nil, err
}


// SubscribeToLeaderboardUpdates registers a callback for leaderboard updates
func (s *playerStatsService) SubscribeToLeaderboardUpdates(callback LeaderboardUpdateCallback) string {
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	s.nextSubID++
	subID := fmt.Sprintf("sub_%d", s.nextSubID)
	s.subscriptions[subID] = callback

	s.logger.Debug("leaderboard subscription added", "subscriptionID", subID)

	return subID
}

// UnsubscribeFromLeaderboardUpdates removes a leaderboard update subscription
func (s *playerStatsService) UnsubscribeFromLeaderboardUpdates(subscriptionID string) {
	s.subscriptionsMu.Lock()
	defer s.subscriptionsMu.Unlock()

	delete(s.subscriptions, subscriptionID)

	s.logger.Debug("leaderboard subscription removed", "subscriptionID", subscriptionID)
}

// NotifyLeaderboardUpdate notifies all subscribers of a leaderboard update
func (s *playerStatsService) NotifyLeaderboardUpdate(ctx context.Context) {
	// Invalidate leaderboard cache
	s.leaderboardCacheMu.Lock()
	s.leaderboardCache = nil
	s.leaderboardCacheMu.Unlock()

	// Get fresh leaderboard data
	leaderboard, err := s.GetLeaderboard(ctx, 10)
	if err != nil {
		s.logger.Warn("failed to get leaderboard for notification", "error", err)
		return
	}

	// Notify all subscribers
	s.subscriptionsMu.RLock()
	subscribers := make([]LeaderboardUpdateCallback, 0, len(s.subscriptions))
	for _, callback := range s.subscriptions {
		subscribers = append(subscribers, callback)
	}
	s.subscriptionsMu.RUnlock()

	for _, callback := range subscribers {
		go func(cb LeaderboardUpdateCallback) {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("panic in leaderboard update callback", "error", r)
				}
			}()
			cb(leaderboard)
		}(callback)
	}

	if len(subscribers) > 0 {
		s.logger.Debug("leaderboard update notified", "subscriberCount", len(subscribers))
	}
}

// InvalidateCache removes a player's stats from the cache
func (s *playerStatsService) InvalidateCache(username string) {
	s.statsCacheMu.Lock()
	delete(s.statsCache, username)
	s.statsCacheMu.Unlock()

	// Also invalidate leaderboard cache since stats changed
	s.leaderboardCacheMu.Lock()
	s.leaderboardCache = nil
	s.leaderboardCacheMu.Unlock()
}

// RefreshLeaderboardCache forces a refresh of the leaderboard cache
func (s *playerStatsService) RefreshLeaderboardCache(ctx context.Context) error {
	// Invalidate current cache
	s.leaderboardCacheMu.Lock()
	s.leaderboardCache = nil
	s.leaderboardCacheMu.Unlock()

	// Fetch fresh data
	_, err := s.GetLeaderboard(ctx, 100)
	return err
}

// GetCacheStats returns statistics about the cache
func (s *playerStatsService) GetCacheStats() map[string]interface{} {
	s.statsCacheMu.RLock()
	statsCacheSize := len(s.statsCache)
	s.statsCacheMu.RUnlock()

	s.leaderboardCacheMu.RLock()
	leaderboardCached := s.leaderboardCache != nil
	leaderboardAge := time.Since(s.leaderboardCachedAt)
	s.leaderboardCacheMu.RUnlock()

	s.subscriptionsMu.RLock()
	subscriptionCount := len(s.subscriptions)
	s.subscriptionsMu.RUnlock()

	return map[string]interface{}{
		"stats_cache_size":      statsCacheSize,
		"leaderboard_cached":    leaderboardCached,
		"leaderboard_age_ms":    leaderboardAge.Milliseconds(),
		"subscription_count":    subscriptionCount,
	}
}

// getCachedStats retrieves stats from cache if valid
func (s *playerStatsService) getCachedStats(username string) *models.PlayerStats {
	s.statsCacheMu.RLock()
	defer s.statsCacheMu.RUnlock()

	cached, ok := s.statsCache[username]
	if !ok {
		return nil
	}

	// Check if cache is still valid
	if time.Since(cached.CachedAt) > s.statsCacheTTL {
		return nil
	}

	return cached.Stats
}

// cacheStats adds stats to the cache
func (s *playerStatsService) cacheStats(stats *models.PlayerStats) {
	if stats == nil {
		return
	}

	s.statsCacheMu.Lock()
	defer s.statsCacheMu.Unlock()

	s.statsCache[stats.Username] = &cachedStats{
		Stats:    stats,
		CachedAt: time.Now(),
	}
}

// CleanupExpiredCache removes expired entries from the cache
func (s *playerStatsService) CleanupExpiredCache() int {
	s.statsCacheMu.Lock()
	defer s.statsCacheMu.Unlock()

	removed := 0
	now := time.Now()

	for username, cached := range s.statsCache {
		if now.Sub(cached.CachedAt) > s.statsCacheTTL {
			delete(s.statsCache, username)
			removed++
		}
	}

	return removed
}
