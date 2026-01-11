package analytics

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/pkg/models"
)

// Producer handles sending events to Kafka
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) *Producer {
	// Configure Kafka writer
	writerConfig := kafka.WriterConfig{
		Brokers:      []string{cfg.BootstrapServers},
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	}

	// Add authentication if provided
	if cfg.APIKey != "" && cfg.APISecret != "" {
		mechanism := plain.Mechanism{
			Username: cfg.APIKey,
			Password: cfg.APISecret,
		}
		
		writerConfig.Dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		}
	}

	writer := kafka.NewWriter(writerConfig)

	return &Producer{writer: writer}
}

// SendEvent sends a game event to Kafka
func (p *Producer) SendEvent(ctx context.Context, event *models.GameEvent) error {
	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(event.GameID),
		Value: data,
		Time:  time.Now(),
	}

	// Send message
	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	return nil
}

// SendGameStarted sends a game started event
func (p *Producer) SendGameStarted(ctx context.Context, gameID, player1, player2 string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-started-%d", gameID, time.Now().Unix()),
		EventType: models.EventGameStarted,
		GameID:    gameID,
		PlayerID:  player1,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"player1": player1,
			"player2": player2,
		},
	}
	return p.SendEvent(ctx, event)
}

// SendGameCompleted sends a game completed event
func (p *Producer) SendGameCompleted(ctx context.Context, gameID, winner, loser string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-completed-%d", gameID, time.Now().Unix()),
		EventType: models.EventGameCompleted,
		GameID:    gameID,
		PlayerID:  winner,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"winner": winner,
			"loser":  loser,
		},
	}
	return p.SendEvent(ctx, event)
}

// SendPlayerJoined sends a player joined event
func (p *Producer) SendPlayerJoined(ctx context.Context, gameID, playerID string) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-joined-%s-%d", gameID, playerID, time.Now().Unix()),
		EventType: models.EventPlayerJoined,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"action": "joined",
		},
	}
	return p.SendEvent(ctx, event)
}

// SendMoveEvent sends a move event
func (p *Producer) SendMoveEvent(ctx context.Context, gameID, playerID string, column, row int) error {
	event := &models.GameEvent{
		ID:        fmt.Sprintf("%s-move-%s-%d", gameID, playerID, time.Now().Unix()),
		EventType: models.EventMoveMade,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"column": column,
			"row":    row,
		},
	}
	return p.SendEvent(ctx, event)
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}