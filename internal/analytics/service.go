package analytics

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"gorm.io/gorm"

	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/pkg/models"
)

// Service handles analytics event processing
type Service struct {
	reader *kafka.Reader
	db     *gorm.DB
}

// NewService creates a new analytics service
func NewService(cfg config.KafkaConfig, db *gorm.DB) (*Service, error) {
	// Configure Kafka reader
	readerConfig := kafka.ReaderConfig{
		Brokers:     []string{cfg.BootstrapServers},
		Topic:       cfg.Topic,
		GroupID:     cfg.ConsumerGroup,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	}

	// Add authentication if provided
	if cfg.APIKey != "" && cfg.APISecret != "" {
		mechanism := plain.Mechanism{
			Username: cfg.APIKey,
			Password: cfg.APISecret,
		}
		
		readerConfig.Dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		}
	}

	reader := kafka.NewReader(readerConfig)

	return &Service{
		reader: reader,
		db:     db,
	}, nil
}

// Start starts the analytics service
func (s *Service) Start(ctx context.Context) error {
	defer s.reader.Close()

	log.Println("Analytics service started, waiting for events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Analytics service stopping...")
			return nil
		default:
			// Read message with timeout
			msg, err := s.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				log.Printf("Consumer error: %v", err)
				continue
			}

			// Process the message
			if err := s.processMessage(&msg); err != nil {
				log.Printf("Failed to process message: %v", err)
			}

			// Commit the message
			if err := s.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (s *Service) processMessage(msg *kafka.Message) error {
	var event models.GameEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Store the event in database
	if err := s.db.Create(&event).Error; err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Process event based on type
	switch event.EventType {
	case models.EventGameCompleted:
		if err := s.processGameCompleted(&event); err != nil {
			log.Printf("Failed to process game completed event: %v", err)
		}
	case models.EventPlayerJoined:
		if err := s.processPlayerJoined(&event); err != nil {
			log.Printf("Failed to process player joined event: %v", err)
		}
	}

	log.Printf("Processed event: %s for game %s", event.EventType, event.GameID)
	return nil
}

// processGameCompleted processes game completion events
func (s *Service) processGameCompleted(event *models.GameEvent) error {
	// Extract winner from metadata
	winner, ok := event.Metadata["winner"].(string)
	if !ok {
		return fmt.Errorf("missing winner in game completed event")
	}

	// Update player statistics
	if winner != "draw" {
		// Update winner stats
		if err := s.updatePlayerStats(winner, true); err != nil {
			return fmt.Errorf("failed to update winner stats: %w", err)
		}

		// Update loser stats (get other player from metadata)
		if loser, ok := event.Metadata["loser"].(string); ok {
			if err := s.updatePlayerStats(loser, false); err != nil {
				return fmt.Errorf("failed to update loser stats: %w", err)
			}
		}
	}

	return nil
}

// processPlayerJoined processes player joined events
func (s *Service) processPlayerJoined(event *models.GameEvent) error {
	// Ensure player stats record exists
	var stats models.PlayerStats
	result := s.db.Where("username = ?", event.PlayerID).First(&stats)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new player stats record
		stats = models.PlayerStats{
			Username:    event.PlayerID,
			GamesPlayed: 0,
			GamesWon:    0,
			WinRate:     0.0,
		}
		if err := s.db.Create(&stats).Error; err != nil {
			return fmt.Errorf("failed to create player stats: %w", err)
		}
	}

	return nil
}

// updatePlayerStats updates player statistics
func (s *Service) updatePlayerStats(username string, won bool) error {
	var stats models.PlayerStats
	result := s.db.Where("username = ?", username).First(&stats)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new stats record
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
		// Update existing stats
		stats.GamesPlayed++
		if won {
			stats.GamesWon++
		}
	}

	// Calculate win rate
	stats.CalculateWinRate()

	// Save updated stats
	if err := s.db.Save(&stats).Error; err != nil {
		return fmt.Errorf("failed to save player stats: %w", err)
	}

	return nil
}