package analytics

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"gorm.io/gorm"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/pkg/models"
)

// ServiceConfig holds configuration for the analytics service
type ServiceConfig struct {
	MaxConcurrentProcessing  int
	ProcessingTimeout        time.Duration
	CommitInterval           time.Duration
	MetricsFlushInterval     time.Duration
	EnableMetricsAggregation bool
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxConcurrentProcessing:  10,
		ProcessingTimeout:        30 * time.Second,
		CommitInterval:           5 * time.Second,
		MetricsFlushInterval:     1 * time.Minute,
		EnableMetricsAggregation: true,
	}
}

// GameMetrics holds aggregated game metrics (Requirement 10.2, 10.3, 10.4)
type GameMetrics struct {
	// Time-based metrics
	GamesCompletedLastHour int64
	GamesCompletedLastDay  int64

	// Duration metrics
	TotalGameDuration   time.Duration
	GameCount           int64
	AverageGameDuration time.Duration
	MinGameDuration     time.Duration
	MaxGameDuration     time.Duration

	// Player metrics
	TotalMoves            int64
	AverageMovesPerGame   float64
	UniquePlayersLastHour map[string]bool

	// Win/loss metrics
	WinsByPlayer map[string]int64

	// Last updated
	LastUpdated time.Time

	mutex sync.RWMutex
}

// NewGameMetrics creates a new GameMetrics instance
func NewGameMetrics() *GameMetrics {
	return &GameMetrics{
		UniquePlayersLastHour: make(map[string]bool),
		WinsByPlayer:          make(map[string]int64),
		MinGameDuration:       time.Duration(1<<63 - 1), // Max duration
		LastUpdated:           time.Now(),
	}
}

// Service handles analytics event processing with enhanced features
type Service struct {
	reader  *kafka.Reader
	db      *gorm.DB
	config  *ServiceConfig
	logger  *slog.Logger
	metrics *GameMetrics

	// Processing state
	eventsProcessed atomic.Int64
	eventsFailed    atomic.Int64
	lastCommitTime  time.Time

	// Shutdown coordination
	shutdownCh   chan struct{}
	processingWg sync.WaitGroup
}

// NewService creates a new analytics service
func NewService(cfg config.KafkaConfig, db *gorm.DB) (*Service, error) {
	return NewServiceWithConfig(cfg, db, DefaultServiceConfig())
}

// NewServiceWithConfig creates a new analytics service with custom configuration
func NewServiceWithConfig(cfg config.KafkaConfig, db *gorm.DB, serviceCfg *ServiceConfig) (*Service, error) {
	logger := slog.Default().With("component", "analytics-service")

	// Create dialer for Confluent Cloud authentication
	var dialer *kafka.Dialer
	if cfg.APIKey != "" && cfg.APISecret != "" {
		mechanism := plain.Mechanism{
			Username: cfg.APIKey,
			Password: cfg.APISecret,
		}

		dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
			TLS:           &tls.Config{MinVersion: tls.VersionTLS12},
		}
	}

	// Configure Kafka reader with consumer group management
	readerConfig := kafka.ReaderConfig{
		Brokers:               []string{cfg.BootstrapServers},
		Topic:                 cfg.Topic,
		GroupID:               cfg.ConsumerGroup,
		StartOffset:           kafka.FirstOffset,
		MinBytes:              1e3,  // 1KB
		MaxBytes:              10e6, // 10MB
		MaxWait:               3 * time.Second,
		CommitInterval:        serviceCfg.CommitInterval,
		HeartbeatInterval:     3 * time.Second,
		SessionTimeout:        30 * time.Second,
		RebalanceTimeout:      60 * time.Second,
		RetentionTime:         24 * time.Hour,
		WatchPartitionChanges: true,
	}

	if dialer != nil {
		readerConfig.Dialer = dialer
	}

	reader := kafka.NewReader(readerConfig)

	service := &Service{
		reader:     reader,
		db:         db,
		config:     serviceCfg,
		logger:     logger,
		metrics:    NewGameMetrics(),
		shutdownCh: make(chan struct{}),
	}

	logger.Info("Analytics service initialized",
		"brokers", cfg.BootstrapServers,
		"topic", cfg.Topic,
		"consumerGroup", cfg.ConsumerGroup,
	)

	return service, nil
}

// Start starts the analytics service (Requirement 10.1)
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting analytics service...")

	// Start metrics flush goroutine if enabled
	if s.config.EnableMetricsAggregation {
		go s.metricsFlushLoop(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Analytics service stopping due to context cancellation...")
			return s.shutdown()
		case <-s.shutdownCh:
			s.logger.Info("Analytics service stopping due to shutdown signal...")
			return s.shutdown()
		default:
			// Read message with timeout
			readCtx, cancel := context.WithTimeout(ctx, s.config.ProcessingTimeout)
			msg, err := s.reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				s.logger.Warn("Consumer fetch error", "error", err)
				continue
			}

			// Process the message
			s.processingWg.Add(1)
			go func(m kafka.Message) {
				defer s.processingWg.Done()

				if err := s.processMessage(ctx, &m); err != nil {
					s.eventsFailed.Add(1)
					s.logger.Error("Failed to process message",
						"error", err,
						"partition", m.Partition,
						"offset", m.Offset,
					)
				} else {
					s.eventsProcessed.Add(1)
				}

				// Commit the message
				if err := s.reader.CommitMessages(ctx, m); err != nil {
					s.logger.Error("Failed to commit message", "error", err)
				}
			}(msg)
		}
	}
}

// metricsFlushLoop periodically flushes aggregated metrics to database
func (s *Service) metricsFlushLoop(ctx context.Context) {
	ticker := time.NewTicker(s.config.MetricsFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdownCh:
			return
		case <-ticker.C:
			if err := s.flushMetrics(ctx); err != nil {
				s.logger.Error("Failed to flush metrics", "error", err)
			}
		}
	}
}

// flushMetrics saves aggregated metrics to database (Requirement 10.5)
func (s *Service) flushMetrics(ctx context.Context) error {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()

	// Create metrics snapshot
	snapshot := &models.AnalyticsSnapshot{
		Timestamp:          time.Now(),
		GamesCompletedHour: s.metrics.GamesCompletedLastHour,
		GamesCompletedDay:  s.metrics.GamesCompletedLastDay,
		AvgGameDurationSec: int64(s.metrics.AverageGameDuration.Seconds()),
		TotalMoves:         s.metrics.TotalMoves,
		UniquePlayersHour:  int64(len(s.metrics.UniquePlayersLastHour)),
	}

	// Save to database
	if err := s.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return fmt.Errorf("failed to save metrics snapshot: %w", err)
	}

	s.logger.Info("Metrics flushed to database",
		"gamesCompletedHour", snapshot.GamesCompletedHour,
		"avgDurationSec", snapshot.AvgGameDurationSec,
		"uniquePlayers", snapshot.UniquePlayersHour,
	)

	// Reset hourly counters
	s.metrics.GamesCompletedLastHour = 0
	s.metrics.UniquePlayersLastHour = make(map[string]bool)
	s.metrics.LastUpdated = time.Now()

	return nil
}

// shutdown gracefully shuts down the service
func (s *Service) shutdown() error {
	// Wait for in-flight processing to complete
	s.processingWg.Wait()

	// Final metrics flush
	if s.config.EnableMetricsAggregation {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.flushMetrics(ctx); err != nil {
			s.logger.Warn("Failed to flush final metrics", "error", err)
		}
	}

	s.logger.Info("Analytics service shutdown complete",
		"eventsProcessed", s.eventsProcessed.Load(),
		"eventsFailed", s.eventsFailed.Load(),
	)

	return s.reader.Close()
}

// Stop signals the service to stop
func (s *Service) Stop() {
	close(s.shutdownCh)
}

// GetStats returns service statistics
func (s *Service) GetStats() map[string]interface{} {
	s.metrics.mutex.RLock()
	defer s.metrics.mutex.RUnlock()

	return map[string]interface{}{
		"eventsProcessed":    s.eventsProcessed.Load(),
		"eventsFailed":       s.eventsFailed.Load(),
		"gamesCompletedHour": s.metrics.GamesCompletedLastHour,
		"gamesCompletedDay":  s.metrics.GamesCompletedLastDay,
		"avgGameDurationSec": s.metrics.AverageGameDuration.Seconds(),
		"totalMoves":         s.metrics.TotalMoves,
		"uniquePlayersHour":  len(s.metrics.UniquePlayersLastHour),
	}
}

// processMessage processes a single Kafka message (Requirement 10.1)
func (s *Service) processMessage(ctx context.Context, msg *kafka.Message) error {
	var event models.GameEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	s.logger.Debug("Processing event",
		"eventType", event.EventType,
		"gameID", event.GameID,
		"playerID", event.PlayerID,
	)

	// Store the event in database
	if err := s.db.WithContext(ctx).Create(&event).Error; err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Update aggregated metrics
	s.updateMetrics(&event)

	// Process event based on type
	switch event.EventType {
	case models.EventGameStarted:
		if err := s.processGameStarted(ctx, &event); err != nil {
			s.logger.Warn("Failed to process game started event", "error", err)
		}
	case models.EventMoveMade:
		if err := s.processMoveMade(ctx, &event); err != nil {
			s.logger.Warn("Failed to process move made event", "error", err)
		}
	case models.EventGameCompleted:
		if err := s.processGameCompleted(ctx, &event); err != nil {
			s.logger.Warn("Failed to process game completed event", "error", err)
		}
	case models.EventPlayerJoined:
		if err := s.processPlayerJoined(ctx, &event); err != nil {
			s.logger.Warn("Failed to process player joined event", "error", err)
		}
	case models.EventPlayerLeft:
		if err := s.processPlayerDisconnected(ctx, &event); err != nil {
			s.logger.Warn("Failed to process player disconnected event", "error", err)
		}
	case models.EventPlayerReconnected:
		if err := s.processPlayerReconnected(ctx, &event); err != nil {
			s.logger.Warn("Failed to process player reconnected event", "error", err)
		}
	}

	s.logger.Debug("Event processed successfully",
		"eventType", event.EventType,
		"gameID", event.GameID,
	)

	return nil
}

// updateMetrics updates the aggregated metrics based on event type
func (s *Service) updateMetrics(event *models.GameEvent) {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()

	// Track unique players
	s.metrics.UniquePlayersLastHour[event.PlayerID] = true

	switch event.EventType {
	case models.EventMoveMade:
		s.metrics.TotalMoves++
	case models.EventGameCompleted:
		s.metrics.GamesCompletedLastHour++
		s.metrics.GamesCompletedLastDay++
		s.metrics.GameCount++

		// Calculate game duration from metadata
		if durationMs, ok := event.Metadata["durationMs"].(float64); ok {
			duration := time.Duration(int64(durationMs)) * time.Millisecond
			s.metrics.TotalGameDuration += duration

			if duration < s.metrics.MinGameDuration {
				s.metrics.MinGameDuration = duration
			}
			if duration > s.metrics.MaxGameDuration {
				s.metrics.MaxGameDuration = duration
			}

			// Recalculate average
			s.metrics.AverageGameDuration = s.metrics.TotalGameDuration / time.Duration(s.metrics.GameCount)
		}

		// Track wins by player
		if winner, ok := event.Metadata["winner"].(string); ok && winner != "" && winner != "draw" {
			s.metrics.WinsByPlayer[winner]++
		}
	}

	s.metrics.LastUpdated = time.Now()
}

// processGameStarted handles game started events
func (s *Service) processGameStarted(ctx context.Context, event *models.GameEvent) error {
	player1, _ := event.Metadata["player1"].(string)
	player2, _ := event.Metadata["player2"].(string)

	s.logger.Info("Game started",
		"gameID", event.GameID,
		"player1", player1,
		"player2", player2,
	)

	// Ensure player stats records exist for both players
	for _, player := range []string{player1, player2} {
		if player == "" {
			continue
		}
		if err := s.ensurePlayerStats(ctx, player); err != nil {
			s.logger.Warn("Failed to ensure player stats", "player", player, "error", err)
		}
	}

	return nil
}

// processMoveMade handles move made events (Requirement 10.2 - timing data)
func (s *Service) processMoveMade(ctx context.Context, event *models.GameEvent) error {
	column, _ := event.Metadata["column"].(float64)
	row, _ := event.Metadata["row"].(float64)
	moveNumber, _ := event.Metadata["moveNumber"].(float64)

	s.logger.Debug("Move made",
		"gameID", event.GameID,
		"player", event.PlayerID,
		"column", int(column),
		"row", int(row),
		"moveNumber", int(moveNumber),
	)

	return nil
}

// processGameCompleted handles game completion events (Requirement 10.2, 10.3)
func (s *Service) processGameCompleted(ctx context.Context, event *models.GameEvent) error {
	winner, _ := event.Metadata["winner"].(string)
	loser, _ := event.Metadata["loser"].(string)
	durationSec, _ := event.Metadata["durationSec"].(float64)

	s.logger.Info("Game completed",
		"gameID", event.GameID,
		"winner", winner,
		"loser", loser,
		"durationSec", int(durationSec),
	)

	// Update player statistics
	if winner != "" && winner != "draw" {
		if err := s.updatePlayerStats(ctx, winner, true); err != nil {
			return fmt.Errorf("failed to update winner stats: %w", err)
		}

		if loser != "" {
			if err := s.updatePlayerStats(ctx, loser, false); err != nil {
				return fmt.Errorf("failed to update loser stats: %w", err)
			}
		}
	}

	return nil
}

// processPlayerJoined handles player joined events
func (s *Service) processPlayerJoined(ctx context.Context, event *models.GameEvent) error {
	s.logger.Debug("Player joined",
		"gameID", event.GameID,
		"player", event.PlayerID,
	)

	return s.ensurePlayerStats(ctx, event.PlayerID)
}

// processPlayerDisconnected handles player disconnected events
func (s *Service) processPlayerDisconnected(ctx context.Context, event *models.GameEvent) error {
	s.logger.Info("Player disconnected",
		"gameID", event.GameID,
		"player", event.PlayerID,
	)
	return nil
}

// processPlayerReconnected handles player reconnected events
func (s *Service) processPlayerReconnected(ctx context.Context, event *models.GameEvent) error {
	s.logger.Info("Player reconnected",
		"gameID", event.GameID,
		"player", event.PlayerID,
	)
	return nil
}

// ensurePlayerStats ensures a player stats record exists
func (s *Service) ensurePlayerStats(ctx context.Context, username string) error {
	if username == "" {
		return nil
	}

	var stats models.PlayerStats
	result := s.db.WithContext(ctx).Where("username = ?", username).First(&stats)

	if result.Error == gorm.ErrRecordNotFound {
		stats = models.PlayerStats{
			Username:    username,
			GamesPlayed: 0,
			GamesWon:    0,
			WinRate:     0.0,
		}
		if err := s.db.WithContext(ctx).Create(&stats).Error; err != nil {
			return fmt.Errorf("failed to create player stats: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to check player stats: %w", result.Error)
	}

	return nil
}

// updatePlayerStats updates player statistics (Requirement 10.3)
func (s *Service) updatePlayerStats(ctx context.Context, username string, won bool) error {
	if username == "" {
		return nil
	}

	var stats models.PlayerStats
	result := s.db.WithContext(ctx).Where("username = ?", username).First(&stats)

	if result.Error == gorm.ErrRecordNotFound {
		stats = models.PlayerStats{
			Username:    username,
			GamesPlayed: 1,
			GamesWon:    0,
		}
		if won {
			stats.GamesWon = 1
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to fetch player stats: %w", result.Error)
	} else {
		stats.GamesPlayed++
		if won {
			stats.GamesWon++
		}
	}

	// Calculate win rate
	stats.CalculateWinRate()

	// Save updated stats
	if err := s.db.WithContext(ctx).Save(&stats).Error; err != nil {
		return fmt.Errorf("failed to save player stats: %w", err)
	}

	s.logger.Debug("Player stats updated",
		"username", username,
		"gamesPlayed", stats.GamesPlayed,
		"gamesWon", stats.GamesWon,
		"winRate", stats.WinRate,
	)

	return nil
}
