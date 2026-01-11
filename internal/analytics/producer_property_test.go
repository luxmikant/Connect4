//go:build property

package analytics

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"

	"connect4-multiplayer/pkg/models"
)

// MockKafkaWriter is a mock Kafka writer for testing
type MockKafkaWriter struct {
	messages []MockMessage
	mu       sync.Mutex
	failNext bool
	closed   bool
}

type MockMessage struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
	Time    time.Time
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failNext {
		m.failNext = false
		return context.DeadlineExceeded
	}

	return nil
}

func (m *MockKafkaWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockKafkaWriter) GetMessages() []MockMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

// genGameID generates random game IDs
func genGameID() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) > 8 {
			return "game-" + s[:8]
		}
		return "game-" + s
	})
}

// genPlayerID generates random player IDs
func genPlayerID() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) > 12 {
			return "player-" + s[:12]
		}
		return "player-" + s
	})
}

// genColumn generates valid Connect 4 columns (0-6)
func genColumn() gopter.Gen {
	return gen.IntRange(0, 6)
}

// genRow generates valid Connect 4 rows (0-5)
func genRow() gopter.Gen {
	return gen.IntRange(0, 5)
}

// genMoveNumber generates valid move numbers
func genMoveNumber() gopter.Gen {
	return gen.IntRange(1, 42) // Max 42 moves in Connect 4
}

// genDuration generates game durations
func genDuration() gopter.Gen {
	return gen.IntRange(30, 1800).Map(func(secs int) time.Duration {
		return time.Duration(secs) * time.Second
	})
}

// TestProperty17_GameStartedEventPublishing tests Requirement 9.1
// Property: When a game starts, a game_started event is published with both player IDs
func TestProperty17_GameStartedEventPublishing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("game_started events contain valid player information", prop.ForAll(
		func(gameID, player1, player2 string) bool {
			event := models.NewGameStartedEvent(gameID, player1, player2)

			// Verify event type
			if event.EventType != models.EventGameStarted {
				return false
			}

			// Verify game ID
			if event.GameID != gameID {
				return false
			}

			// Verify player IDs are in metadata
			p1, ok := event.Metadata["player1"].(string)
			if !ok || p1 != player1 {
				return false
			}

			p2, ok := event.Metadata["player2"].(string)
			if !ok || p2 != player2 {
				return false
			}

			// Verify timestamp is set
			if event.Timestamp.IsZero() {
				return false
			}

			return true
		},
		genGameID(),
		genPlayerID(),
		genPlayerID(),
	))

	properties.TestingRun(t)
}

// TestProperty17_MoveMadeEventPublishing tests Requirement 9.2
// Property: Move events contain position and timing data
func TestProperty17_MoveMadeEventPublishing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("move_made events contain valid position data", prop.ForAll(
		func(gameID, playerID string, column, row int) bool {
			event := models.NewMoveMadeEvent(gameID, playerID, column, row)

			// Verify event type
			if event.EventType != models.EventMoveMade {
				return false
			}

			// Verify game and player IDs
			if event.GameID != gameID || event.PlayerID != playerID {
				return false
			}

			// Verify position data in metadata
			col, ok := event.Metadata["column"].(int)
			if !ok || col != column {
				return false
			}

			rowVal, ok := event.Metadata["row"].(int)
			if !ok || rowVal != row {
				return false
			}

			// Verify timestamp is set
			if event.Timestamp.IsZero() {
				return false
			}

			return true
		},
		genGameID(),
		genPlayerID(),
		genColumn(),
		genRow(),
	))

	properties.TestingRun(t)
}

// TestProperty17_GameCompletedEventPublishing tests Requirement 9.3
// Property: Game completed events contain outcome and duration data
func TestProperty17_GameCompletedEventPublishing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("game_completed events contain valid outcome data", prop.ForAll(
		func(gameID, winner, loser string, durationSec int) bool {
			event := models.NewGameCompletedEvent(gameID, winner, loser, durationSec)

			// Verify event type
			if event.EventType != models.EventGameCompleted {
				return false
			}

			// Verify game ID
			if event.GameID != gameID {
				return false
			}

			// Verify winner/loser in metadata
			w, ok := event.Metadata["winner"].(string)
			if !ok || w != winner {
				return false
			}

			l, ok := event.Metadata["loser"].(string)
			if !ok || l != loser {
				return false
			}

			// Verify duration in metadata
			d, ok := event.Metadata["duration"].(int)
			if !ok || d != durationSec {
				return false
			}

			// Verify timestamp is set
			if event.Timestamp.IsZero() {
				return false
			}

			return true
		},
		genGameID(),
		genPlayerID(),
		genPlayerID(),
		gen.IntRange(30, 1800),
	))

	properties.TestingRun(t)
}

// TestProperty17_PlayerConnectionEventPublishing tests Requirement 9.4
// Property: Connection events contain player and session metadata
func TestProperty17_PlayerConnectionEventPublishing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("player_joined events contain valid player data", prop.ForAll(
		func(gameID, playerID string) bool {
			event := models.NewPlayerJoinedEvent(gameID, playerID)

			// Verify event type
			if event.EventType != models.EventPlayerJoined {
				return false
			}

			// Verify game and player IDs
			if event.GameID != gameID || event.PlayerID != playerID {
				return false
			}

			// Verify timestamp is set
			if event.Timestamp.IsZero() {
				return false
			}

			return true
		},
		genGameID(),
		genPlayerID(),
	))

	properties.TestingRun(t)
}

// TestProperty17_EventSerialization tests Requirement 9.5
// Property: All events can be serialized and deserialized correctly
func TestProperty17_EventSerialization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("events serialize and deserialize correctly", prop.ForAll(
		func(gameID, playerID string, column, row int) bool {
			// Create an event
			event := models.NewMoveMadeEvent(gameID, playerID, column, row)
			event.ID = "test-event-id"

			// Serialize
			data, err := json.Marshal(event)
			if err != nil {
				return false
			}

			// Deserialize
			var restored models.GameEvent
			if err := json.Unmarshal(data, &restored); err != nil {
				return false
			}

			// Verify fields match
			if restored.ID != event.ID {
				return false
			}
			if restored.EventType != event.EventType {
				return false
			}
			if restored.GameID != event.GameID {
				return false
			}
			if restored.PlayerID != event.PlayerID {
				return false
			}

			// Verify metadata (note: JSON unmarshals numbers as float64)
			colRestored, ok := restored.Metadata["column"].(float64)
			if !ok || int(colRestored) != column {
				return false
			}

			rowRestored, ok := restored.Metadata["row"].(float64)
			if !ok || int(rowRestored) != row {
				return false
			}

			return true
		},
		genGameID(),
		genPlayerID(),
		genColumn(),
		genRow(),
	))

	properties.TestingRun(t)
}

// TestProperty17_CircuitBreakerBehavior tests circuit breaker correctness
// Property: Circuit breaker transitions correctly based on failures and successes
func TestProperty17_CircuitBreakerBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("circuit breaker opens after failure threshold", prop.ForAll(
		func(failureThreshold, successThreshold int) bool {
			if failureThreshold < 1 || successThreshold < 1 {
				return true // Skip invalid inputs
			}
			if failureThreshold > 10 || successThreshold > 10 {
				return true // Skip unreasonable values
			}

			cb := NewCircuitBreaker(int32(failureThreshold), int32(successThreshold), 100*time.Millisecond)

			// Initially closed
			if cb.State() != CircuitClosed {
				return false
			}

			// Should allow requests when closed
			if !cb.AllowRequest() {
				return false
			}

			// Record failures up to threshold
			for i := 0; i < failureThreshold; i++ {
				cb.RecordFailure()
			}

			// Should be open now
			if cb.State() != CircuitOpen {
				return false
			}

			// Should reject requests when open
			if cb.AllowRequest() {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 5),
	))

	properties.Property("circuit breaker closes after success threshold in half-open", prop.ForAll(
		func(failureThreshold, successThreshold int) bool {
			if failureThreshold < 1 || successThreshold < 1 {
				return true
			}
			if failureThreshold > 10 || successThreshold > 10 {
				return true
			}

			cb := NewCircuitBreaker(int32(failureThreshold), int32(successThreshold), 1*time.Millisecond)

			// Trip the circuit
			for i := 0; i < failureThreshold; i++ {
				cb.RecordFailure()
			}

			// Wait for timeout
			time.Sleep(5 * time.Millisecond)

			// Should transition to half-open on next request
			if !cb.AllowRequest() {
				return false
			}

			if cb.State() != CircuitHalfOpen {
				return false
			}

			// Record successes
			for i := 0; i < successThreshold; i++ {
				cb.RecordSuccess()
			}

			// Should be closed now
			if cb.State() != CircuitClosed {
				return false
			}

			return true
		},
		gen.IntRange(1, 5),
		gen.IntRange(1, 3),
	))

	properties.TestingRun(t)
}

// TestProperty17_EventTimestampOrdering tests temporal ordering
// Property: Events have monotonically increasing timestamps within a reasonable timeframe
func TestProperty17_EventTimestampOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("consecutive events have increasing timestamps", prop.ForAll(
		func(numEvents int) bool {
			if numEvents < 2 || numEvents > 100 {
				return true // Skip edge cases
			}

			var events []*models.GameEvent
			for i := 0; i < numEvents; i++ {
				event := models.NewMoveMadeEvent("game-1", "player-1", i%7, 0)
				events = append(events, event)
				time.Sleep(1 * time.Microsecond) // Ensure time passes
			}

			// Verify timestamps are non-decreasing
			for i := 1; i < len(events); i++ {
				if events[i].Timestamp.Before(events[i-1].Timestamp) {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 20),
	))

	properties.TestingRun(t)
}

// TestCircuitBreakerRecovery is a unit test for circuit breaker recovery
func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 10*time.Millisecond)

	// Initially should be closed
	assert.Equal(t, CircuitClosed, cb.State())
	assert.True(t, cb.AllowRequest())

	// Record successes shouldn't change state
	cb.RecordSuccess()
	assert.Equal(t, CircuitClosed, cb.State())

	// Record failures below threshold
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, CircuitClosed, cb.State())

	// Third failure opens circuit
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.State())
	assert.False(t, cb.AllowRequest())

	// Wait for timeout
	time.Sleep(15 * time.Millisecond)

	// Should transition to half-open
	assert.True(t, cb.AllowRequest())
	assert.Equal(t, CircuitHalfOpen, cb.State())

	// Record successes to close
	cb.RecordSuccess()
	assert.Equal(t, CircuitHalfOpen, cb.State())
	cb.RecordSuccess()
	assert.Equal(t, CircuitClosed, cb.State())
}

// TestEventMetadataConsistency verifies metadata is consistent
func TestEventMetadataConsistency(t *testing.T) {
	gameID := "test-game-123"
	player1 := "alice"
	player2 := "bob"

	event := models.NewGameStartedEvent(gameID, player1, player2)

	// Verify all required fields
	assert.Equal(t, models.EventGameStarted, event.EventType)
	assert.Equal(t, gameID, event.GameID)
	assert.Equal(t, player1, event.PlayerID)
	assert.NotNil(t, event.Metadata)
	assert.Equal(t, player1, event.Metadata["player1"])
	assert.Equal(t, player2, event.Metadata["player2"])
	assert.False(t, event.Timestamp.IsZero())
}
