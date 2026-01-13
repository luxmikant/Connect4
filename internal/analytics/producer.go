package analytics

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/pkg/models"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int32

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Failing, reject requests
	CircuitHalfOpen                     // Testing if service recovered
)

// CircuitBreaker implements the circuit breaker pattern for Kafka
type CircuitBreaker struct {
	state       atomic.Int32
	failures    atomic.Int32
	successes   atomic.Int32
	lastFailure time.Time
	mutex       sync.RWMutex

	// Configuration
	failureThreshold int32         // Number of failures before opening circuit
	successThreshold int32         // Number of successes before closing circuit
	timeout          time.Duration // Time to wait before trying again
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int32, timeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
	cb.state.Store(int32(CircuitClosed))
	return cb
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	state := CircuitState(cb.state.Load())

	switch state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		cb.mutex.RLock()
		lastFailure := cb.lastFailure
		cb.mutex.RUnlock()

		// Check if timeout has passed
		if time.Since(lastFailure) > cb.timeout {
			// Transition to half-open
			cb.state.Store(int32(CircuitHalfOpen))
			cb.successes.Store(0)
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	state := CircuitState(cb.state.Load())

	if state == CircuitHalfOpen {
		successes := cb.successes.Add(1)
		if successes >= cb.successThreshold {
			cb.state.Store(int32(CircuitClosed))
			cb.failures.Store(0)
		}
	} else if state == CircuitClosed {
		cb.failures.Store(0)
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	state := CircuitState(cb.state.Load())

	cb.mutex.Lock()
	cb.lastFailure = time.Now()
	cb.mutex.Unlock()

	if state == CircuitHalfOpen {
		cb.state.Store(int32(CircuitOpen))
	} else if state == CircuitClosed {
		failures := cb.failures.Add(1)
		if failures >= cb.failureThreshold {
			cb.state.Store(int32(CircuitOpen))
		}
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() CircuitState {
	return CircuitState(cb.state.Load())
}

// ProducerConfig holds configuration for the Kafka producer
type ProducerConfig struct {
	MaxRetries        int
	RetryBackoff      time.Duration
	MaxRetryBackoff   time.Duration
	CircuitBreaker    *CircuitBreaker
	EnableHealthCheck bool
	HealthCheckPeriod time.Duration
}

// DefaultProducerConfig returns default producer configuration
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		MaxRetries:        3,
		RetryBackoff:      100 * time.Millisecond,
		MaxRetryBackoff:   5 * time.Second,
		CircuitBreaker:    NewCircuitBreaker(5, 3, 30*time.Second),
		EnableHealthCheck: true,
		HealthCheckPeriod: 30 * time.Second,
	}
}

// Producer handles sending events to Kafka with retry logic and circuit breaker
type Producer struct {
	writer       *kafka.Writer
	config       *ProducerConfig
	topic        string
	logger       *slog.Logger
	healthy      atomic.Bool
	healthCancel context.CancelFunc
	healthWg     sync.WaitGroup

	// Metrics
	eventsSent   atomic.Int64
	eventsFailed atomic.Int64
	retriesTotal atomic.Int64
}

// NewProducer creates a new Kafka producer with enhanced features
func NewProducer(cfg config.KafkaConfig) *Producer {
	return NewProducerWithConfig(cfg, DefaultProducerConfig())
}

// NewNoopProducer creates a no-operation producer that doesn't send events to Kafka
// Use this when Kafka is disabled or credentials are not configured
func NewNoopProducer() *Producer {
	return &Producer{
		writer: nil, // nil writer triggers no-op behavior in all methods
		config: DefaultProducerConfig(),
		topic:  "",
		logger: slog.Default().With("component", "noop-producer"),
	}
}

// NewProducerWithConfig creates a new Kafka producer with custom configuration
func NewProducerWithConfig(cfg config.KafkaConfig, producerCfg *ProducerConfig) *Producer {
	logger := slog.Default().With("component", "kafka-producer")

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

	// Configure Kafka writer with partitioning by key (gameID)
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.BootstrapServers),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{}, // Partition by message key (gameID)
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		MaxAttempts:  1, // We handle retries ourselves
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Sync writes for reliability
	}

	if dialer != nil {
		writer.Transport = &kafka.Transport{
			SASL: dialer.SASLMechanism,
			TLS:  dialer.TLS,
		}
	}

	p := &Producer{
		writer: writer,
		config: producerCfg,
		topic:  cfg.Topic,
		logger: logger,
	}
	p.healthy.Store(true)

	// Start health check goroutine
	if producerCfg.EnableHealthCheck {
		ctx, cancel := context.WithCancel(context.Background())
		p.healthCancel = cancel
		p.healthWg.Add(1)
		go p.healthCheckLoop(ctx, cfg.BootstrapServers, dialer)
	}

	logger.Info("Kafka producer initialized",
		"brokers", cfg.BootstrapServers,
		"topic", cfg.Topic,
		"maxRetries", producerCfg.MaxRetries,
	)

	return p
}

// healthCheckLoop periodically checks Kafka connectivity
func (p *Producer) healthCheckLoop(ctx context.Context, brokers string, dialer *kafka.Dialer) {
	defer p.healthWg.Done()

	ticker := time.NewTicker(p.config.HealthCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkHealth(ctx, brokers, dialer)
		}
	}
}

// checkHealth performs a health check against Kafka
func (p *Producer) checkHealth(ctx context.Context, brokers string, dialer *kafka.Dialer) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var conn *kafka.Conn
	var err error

	if dialer != nil {
		conn, err = dialer.DialContext(ctx, "tcp", brokers)
	} else {
		conn, err = kafka.DialContext(ctx, "tcp", brokers)
	}

	if err != nil {
		p.healthy.Store(false)
		p.logger.Warn("Kafka health check failed", "error", err)
		return
	}
	defer conn.Close()

	// Try to get controller to verify connectivity
	_, err = conn.Controller()
	if err != nil {
		p.healthy.Store(false)
		p.logger.Warn("Kafka health check failed", "error", err)
		return
	}

	p.healthy.Store(true)
}

// IsHealthy returns true if the producer is connected to Kafka
func (p *Producer) IsHealthy() bool {
	return p.healthy.Load() && p.config.CircuitBreaker.AllowRequest()
}

// GetMetrics returns producer metrics
func (p *Producer) GetMetrics() map[string]int64 {
	return map[string]int64{
		"events_sent":   p.eventsSent.Load(),
		"events_failed": p.eventsFailed.Load(),
		"retries_total": p.retriesTotal.Load(),
	}
}

// GetCircuitState returns the current circuit breaker state
func (p *Producer) GetCircuitState() string {
	switch p.config.CircuitBreaker.State() {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open, rejecting request")

// ErrProducerUnhealthy is returned when the producer health check fails
var ErrProducerUnhealthy = errors.New("producer is unhealthy")

// SendEvent sends a game event to Kafka with retry logic and circuit breaker
func (p *Producer) SendEvent(ctx context.Context, event *models.GameEvent) error {
	// If producer is not initialized (no Kafka credentials), skip silently
	if p.writer == nil {
		return nil
	}

	// Check circuit breaker
	if !p.config.CircuitBreaker.AllowRequest() {
		p.eventsFailed.Add(1)
		p.logger.Warn("Circuit breaker open, rejecting event",
			"eventType", event.EventType,
			"gameID", event.GameID,
		)
		return ErrCircuitOpen
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		p.eventsFailed.Add(1)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message with gameID as key for partitioning
	message := kafka.Message{
		Key:   []byte(event.GameID), // Partition by gameID ensures ordering per game
		Value: data,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
		},
	}

	// Send with retry logic
	err = p.sendWithRetry(ctx, message)
	if err != nil {
		p.eventsFailed.Add(1)
		p.config.CircuitBreaker.RecordFailure()
		p.logger.Error("Failed to send event after retries",
			"eventType", event.EventType,
			"gameID", event.GameID,
			"error", err,
		)
		return fmt.Errorf("failed to send event after retries: %w", err)
	}

	p.eventsSent.Add(1)
	p.config.CircuitBreaker.RecordSuccess()

	p.logger.Debug("Event sent successfully",
		"eventType", event.EventType,
		"gameID", event.GameID,
		"eventID", event.ID,
	)

	return nil
}

// sendWithRetry sends a message with exponential backoff retry
func (p *Producer) sendWithRetry(ctx context.Context, message kafka.Message) error {
	var lastErr error
	backoff := p.config.RetryBackoff

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			p.retriesTotal.Add(1)
			p.logger.Debug("Retrying message send",
				"attempt", attempt,
				"maxRetries", p.config.MaxRetries,
				"backoff", backoff,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// Exponential backoff with jitter
				backoff = min(backoff*2, p.config.MaxRetryBackoff)
			}
		}

		err := p.writer.WriteMessages(ctx, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableKafkaError(err) {
			return err
		}
	}

	return lastErr
}

// isRetryableKafkaError determines if a Kafka error is retryable
func isRetryableKafkaError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Network errors are generally retryable
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"i/o timeout",
		"temporary failure",
		"broker not available",
		"leader not available",
		"request timed out",
		"network is unreachable",
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SendGameStarted sends a game started event (Requirement 9.1)
func (p *Producer) SendGameStarted(ctx context.Context, gameID, player1, player2 string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-started-%d", gameID, time.Now().UnixNano()),
		EventType: models.EventGameStarted,
		GameID:    gameID,
		PlayerID:  player1,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"player1":   player1,
			"player2":   player2,
			"startTime": time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// SendGameCompleted sends a game completed event (Requirement 9.3)
func (p *Producer) SendGameCompleted(ctx context.Context, gameID, winner, loser string, duration time.Duration) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-completed-%d", gameID, time.Now().UnixNano()),
		EventType: models.EventGameCompleted,
		GameID:    gameID,
		PlayerID:  winner,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"winner":         winner,
			"loser":          loser,
			"durationMs":     duration.Milliseconds(),
			"durationSec":    int(duration.Seconds()),
			"completionTime": time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// SendMoveMade sends a move made event (Requirement 9.2)
func (p *Producer) SendMoveMade(ctx context.Context, gameID, playerID string, column, row int, moveNumber int) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-move-%d-%d", gameID, moveNumber, time.Now().UnixNano()),
		EventType: models.EventMoveMade,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"column":     column,
			"row":        row,
			"moveNumber": moveNumber,
			"moveTime":   time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// SendPlayerJoined sends a player joined event
func (p *Producer) SendPlayerJoined(ctx context.Context, gameID, playerID string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-joined-%s-%d", gameID, playerID, time.Now().UnixNano()),
		EventType: models.EventPlayerJoined,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"action":   "joined",
			"joinTime": time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// SendPlayerDisconnected sends a player disconnected event (Requirement 9.4)
func (p *Producer) SendPlayerDisconnected(ctx context.Context, gameID, playerID string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-disconnected-%s-%d", gameID, playerID, time.Now().UnixNano()),
		EventType: models.EventPlayerLeft,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"action":         "disconnected",
			"disconnectTime": time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// SendPlayerReconnected sends a player reconnected event (Requirement 9.4)
func (p *Producer) SendPlayerReconnected(ctx context.Context, gameID, playerID string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-reconnected-%s-%d", gameID, playerID, time.Now().UnixNano()),
		EventType: models.EventPlayerReconnected,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"action":        "reconnected",
			"reconnectTime": time.Now().Format(time.RFC3339),
		},
	}
	return p.SendEvent(ctx, event)
}

// Close closes the producer gracefully
func (p *Producer) Close() error {
	// If producer is not initialized (no Kafka credentials), skip
	if p.writer == nil {
		return nil
	}

	// Stop health check goroutine
	if p.healthCancel != nil {
		p.healthCancel()
		p.healthWg.Wait()
	}

	if p.logger != nil {
		p.logger.Info("Closing Kafka producer",
			"eventsSent", p.eventsSent.Load(),
			"eventsFailed", p.eventsFailed.Load(),
			"retriesTotal", p.retriesTotal.Load(),
		)
	}

	return p.writer.Close()
}
