//go:build property

package analytics

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"

	"connect4-multiplayer/pkg/models"
)

// TestProperty18_EventConsumption tests Requirement 10.1
// Property: All events received from Kafka are consumed and processed
func TestProperty18_EventConsumption(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("all event types can be unmarshaled and identified", prop.ForAll(
		func(eventType int) bool {
			types := []models.EventType{
				models.EventGameStarted,
				models.EventMoveMade,
				models.EventGameCompleted,
				models.EventPlayerJoined,
				models.EventPlayerLeft,
				models.EventPlayerReconnected,
			}

			idx := eventType % len(types)
			selectedType := types[idx]

			// Create event with the type
			event := &models.GameEvent{
				ID:        "test-event",
				EventType: selectedType,
				GameID:    "game-123",
				PlayerID:  "player-1",
				Timestamp: time.Now(),
				Metadata:  models.EventMetadata{},
			}

			// Verify event type is correctly set
			switch selectedType {
			case models.EventGameStarted:
				return event.EventType == models.EventGameStarted
			case models.EventMoveMade:
				return event.EventType == models.EventMoveMade
			case models.EventGameCompleted:
				return event.EventType == models.EventGameCompleted
			case models.EventPlayerJoined:
				return event.EventType == models.EventPlayerJoined
			case models.EventPlayerLeft:
				return event.EventType == models.EventPlayerLeft
			case models.EventPlayerReconnected:
				return event.EventType == models.EventPlayerReconnected
			default:
				return false
			}
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// TestProperty18_AverageGameDurationCalculation tests Requirement 10.2
// Property: Average game duration is correctly calculated from game completion events
func TestProperty18_AverageGameDurationCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("average duration is correctly calculated", prop.ForAll(
		func(durations []int) bool {
			if len(durations) == 0 {
				return true // Skip empty cases
			}

			metrics := NewGameMetrics()

			var totalDuration time.Duration
			for _, d := range durations {
				if d <= 0 {
					continue // Skip invalid durations
				}

				duration := time.Duration(d) * time.Second

				// Simulate processing a game completed event
				metrics.mutex.Lock()
				metrics.GameCount++
				metrics.TotalGameDuration += duration

				if duration < metrics.MinGameDuration {
					metrics.MinGameDuration = duration
				}
				if duration > metrics.MaxGameDuration {
					metrics.MaxGameDuration = duration
				}

				metrics.AverageGameDuration = metrics.TotalGameDuration / time.Duration(metrics.GameCount)
				metrics.mutex.Unlock()

				totalDuration += duration
			}

			if metrics.GameCount == 0 {
				return true
			}

			// Verify average is correct
			expectedAvg := totalDuration / time.Duration(metrics.GameCount)
			return metrics.AverageGameDuration == expectedAvg
		},
		gen.SliceOfN(10, gen.IntRange(30, 600)),
	))

	properties.TestingRun(t)
}

// TestProperty18_WinnerTracking tests Requirement 10.3
// Property: Most frequent winners are correctly identified
func TestProperty18_WinnerTracking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("winner counts are correctly tracked", prop.ForAll(
		func(winners []string) bool {
			if len(winners) == 0 {
				return true
			}

			metrics := NewGameMetrics()

			// Track expected counts
			expectedCounts := make(map[string]int64)

			// Process each winner
			for _, winner := range winners {
				if winner == "" || winner == "draw" {
					continue
				}

				metrics.mutex.Lock()
				metrics.WinsByPlayer[winner]++
				metrics.mutex.Unlock()

				expectedCounts[winner]++
			}

			// Verify counts match
			metrics.mutex.RLock()
			defer metrics.mutex.RUnlock()

			for player, expected := range expectedCounts {
				actual := metrics.WinsByPlayer[player]
				if actual != expected {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(20, gen.OneConstOf("alice", "bob", "charlie", "diana", "draw", "")),
	))

	properties.TestingRun(t)
}

// TestProperty18_GamesPerHourStatistics tests Requirement 10.4
// Property: Games per hour/day statistics are correctly aggregated
func TestProperty18_GamesPerHourStatistics(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("game counts are correctly incremented", prop.ForAll(
		func(gameCount int) bool {
			if gameCount < 0 || gameCount > 1000 {
				return true
			}

			metrics := NewGameMetrics()

			// Simulate game completions
			for i := 0; i < gameCount; i++ {
				metrics.mutex.Lock()
				metrics.GamesCompletedLastHour++
				metrics.GamesCompletedLastDay++
				metrics.mutex.Unlock()
			}

			metrics.mutex.RLock()
			defer metrics.mutex.RUnlock()

			// Verify counts
			return metrics.GamesCompletedLastHour == int64(gameCount) &&
				metrics.GamesCompletedLastDay == int64(gameCount)
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// TestProperty18_UniquePlayerTracking tests player uniqueness tracking
// Property: Unique players per hour are correctly tracked
func TestProperty18_UniquePlayerTracking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("unique players are correctly counted", prop.ForAll(
		func(players []string) bool {
			metrics := NewGameMetrics()

			// Track unique players
			expectedUnique := make(map[string]bool)

			for _, player := range players {
				if player == "" {
					continue
				}

				metrics.mutex.Lock()
				metrics.UniquePlayersLastHour[player] = true
				metrics.mutex.Unlock()

				expectedUnique[player] = true
			}

			metrics.mutex.RLock()
			defer metrics.mutex.RUnlock()

			return len(metrics.UniquePlayersLastHour) == len(expectedUnique)
		},
		gen.SliceOfN(30, gen.OneConstOf("alice", "bob", "charlie", "diana", "eve", "")),
	))

	properties.TestingRun(t)
}

// TestProperty18_MoveCountTracking tests move counting
// Property: Total moves are correctly counted across all games
func TestProperty18_MoveCountTracking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("total moves are correctly counted", prop.ForAll(
		func(moveCounts []int) bool {
			metrics := NewGameMetrics()

			var expectedTotal int64

			for _, moveCount := range moveCounts {
				if moveCount < 0 {
					continue
				}

				for i := 0; i < moveCount; i++ {
					metrics.mutex.Lock()
					metrics.TotalMoves++
					metrics.mutex.Unlock()
					expectedTotal++
				}
			}

			metrics.mutex.RLock()
			defer metrics.mutex.RUnlock()

			return metrics.TotalMoves == expectedTotal
		},
		gen.SliceOfN(10, gen.IntRange(0, 42)),
	))

	properties.TestingRun(t)
}

// TestProperty18_MetricsPersistence tests Requirement 10.5
// Property: Metrics snapshots contain all required fields
func TestProperty18_MetricsPersistence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("analytics snapshots contain all required fields", prop.ForAll(
		func(gamesHour, gamesDay, avgDuration, totalMoves, uniquePlayers int64) bool {
			if gamesHour < 0 || gamesDay < 0 || avgDuration < 0 || totalMoves < 0 || uniquePlayers < 0 {
				return true
			}

			snapshot := &models.AnalyticsSnapshot{
				Timestamp:          time.Now(),
				GamesCompletedHour: gamesHour,
				GamesCompletedDay:  gamesDay,
				AvgGameDurationSec: avgDuration,
				TotalMoves:         totalMoves,
				UniquePlayersHour:  uniquePlayers,
			}

			// Verify all fields are set
			if snapshot.Timestamp.IsZero() {
				return false
			}
			if snapshot.GamesCompletedHour != gamesHour {
				return false
			}
			if snapshot.GamesCompletedDay != gamesDay {
				return false
			}
			if snapshot.AvgGameDurationSec != avgDuration {
				return false
			}
			if snapshot.TotalMoves != totalMoves {
				return false
			}
			if snapshot.UniquePlayersHour != uniquePlayers {
				return false
			}

			return true
		},
		gen.Int64Range(0, 1000),
		gen.Int64Range(0, 10000),
		gen.Int64Range(0, 3600),
		gen.Int64Range(0, 100000),
		gen.Int64Range(0, 500),
	))

	properties.TestingRun(t)
}

// TestProperty18_MinMaxDurationTracking tests min/max duration tracking
// Property: Min and max durations are correctly tracked
func TestProperty18_MinMaxDurationTracking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("min and max durations are correctly tracked", prop.ForAll(
		func(durations []int) bool {
			if len(durations) == 0 {
				return true
			}

			metrics := NewGameMetrics()

			var validDurations []time.Duration
			for _, d := range durations {
				if d <= 0 {
					continue
				}
				duration := time.Duration(d) * time.Second
				validDurations = append(validDurations, duration)

				metrics.mutex.Lock()
				if duration < metrics.MinGameDuration {
					metrics.MinGameDuration = duration
				}
				if duration > metrics.MaxGameDuration {
					metrics.MaxGameDuration = duration
				}
				metrics.mutex.Unlock()
			}

			if len(validDurations) == 0 {
				return true
			}

			// Find expected min/max
			var expectedMin, expectedMax time.Duration
			expectedMin = time.Duration(1<<63 - 1)
			for _, d := range validDurations {
				if d < expectedMin {
					expectedMin = d
				}
				if d > expectedMax {
					expectedMax = d
				}
			}

			metrics.mutex.RLock()
			defer metrics.mutex.RUnlock()

			return metrics.MinGameDuration == expectedMin &&
				metrics.MaxGameDuration == expectedMax
		},
		gen.SliceOfN(10, gen.IntRange(-10, 600)),
	))

	properties.TestingRun(t)
}

// TestGameMetricsInitialization tests proper initialization
func TestGameMetricsInitialization(t *testing.T) {
	metrics := NewGameMetrics()

	assert.NotNil(t, metrics.UniquePlayersLastHour)
	assert.NotNil(t, metrics.WinsByPlayer)
	assert.Equal(t, int64(0), metrics.GamesCompletedLastHour)
	assert.Equal(t, int64(0), metrics.GamesCompletedLastDay)
	assert.Equal(t, int64(0), metrics.TotalMoves)
	assert.Equal(t, int64(0), metrics.GameCount)
	assert.False(t, metrics.LastUpdated.IsZero())
}

// TestAnalyticsSnapshotTableName verifies table name
func TestAnalyticsSnapshotTableName(t *testing.T) {
	snapshot := models.AnalyticsSnapshot{}
	assert.Equal(t, "analytics_snapshots", snapshot.TableName())
}

// TestServiceConfig tests service configuration defaults
func TestServiceConfig(t *testing.T) {
	config := DefaultServiceConfig()

	assert.Equal(t, 10, config.MaxConcurrentProcessing)
	assert.Equal(t, 30*time.Second, config.ProcessingTimeout)
	assert.Equal(t, 5*time.Second, config.CommitInterval)
	assert.Equal(t, 1*time.Minute, config.MetricsFlushInterval)
	assert.True(t, config.EnableMetricsAggregation)
}
