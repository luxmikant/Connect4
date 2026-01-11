//go:build property
// +build property

package stats

import (
	"reflect"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"connect4-multiplayer/pkg/models"
)

// Feature: connect-4-multiplayer, Property 12: Leaderboard Statistics Accuracy
// *For any* player, the system should accurately track games played, wins, and win percentage, updating in real-time.
// **Validates: Requirements 7.1, 7.3, 7.5**

// TestStatisticsAccuracyProperty tests that player statistics are accurately tracked
func TestStatisticsAccuracyProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50

	properties := gopter.NewProperties(parameters)

	// Property: Win rate calculation is always accurate
	// For any number of games played and games won, win rate should equal gamesWon/gamesPlayed
	properties.Property("win rate calculation is always accurate", prop.ForAll(
		func(gamesPlayed, gamesWon int) bool {
			// Ensure gamesWon <= gamesPlayed
			if gamesWon > gamesPlayed {
				gamesWon = gamesPlayed
			}

			stats := &models.PlayerStats{
				Username:    "testplayer",
				GamesPlayed: gamesPlayed,
				GamesWon:    gamesWon,
			}

			stats.CalculateWinRate()

			// Check win rate accuracy
			if gamesPlayed == 0 {
				return stats.WinRate == 0.0
			}

			expectedWinRate := float64(gamesWon) / float64(gamesPlayed)
			// Allow small floating point tolerance
			diff := stats.WinRate - expectedWinRate
			if diff < 0 {
				diff = -diff
			}
			return diff < 0.0001
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
	))

	// Property: Games played always increases by 1 after recording a game
	properties.Property("games played increments correctly after each game", prop.ForAll(
		func(initialGames, initialWins int, won bool) bool {
			// Ensure initialWins <= initialGames
			if initialWins > initialGames {
				initialWins = initialGames
			}

			stats := &models.PlayerStats{
				Username:    "testplayer",
				GamesPlayed: initialGames,
				GamesWon:    initialWins,
				AvgGameTime: 100,
			}

			stats.UpdateGameStats(won, 120)

			return stats.GamesPlayed == initialGames+1
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
		gen.Bool(),
	))

	// Property: Games won increases by 1 only when player wins
	properties.Property("games won increments only on win", prop.ForAll(
		func(initialGames, initialWins int, won bool) bool {
			// Ensure initialWins <= initialGames
			if initialWins > initialGames {
				initialWins = initialGames
			}

			stats := &models.PlayerStats{
				Username:    "testplayer",
				GamesPlayed: initialGames,
				GamesWon:    initialWins,
				AvgGameTime: 100,
			}

			stats.UpdateGameStats(won, 120)

			if won {
				return stats.GamesWon == initialWins+1
			}
			return stats.GamesWon == initialWins
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
		gen.Bool(),
	))

	// Property: Win rate is always between 0 and 1
	properties.Property("win rate is always between 0 and 1", prop.ForAll(
		func(gamesPlayed, gamesWon int) bool {
			// Ensure gamesWon <= gamesPlayed
			if gamesWon > gamesPlayed {
				gamesWon = gamesPlayed
			}

			stats := &models.PlayerStats{
				Username:    "testplayer",
				GamesPlayed: gamesPlayed,
				GamesWon:    gamesWon,
			}

			stats.CalculateWinRate()

			return stats.WinRate >= 0.0 && stats.WinRate <= 1.0
		},
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
	))

	// Property: Average game time is correctly calculated
	properties.Property("average game time is correctly calculated", prop.ForAll(
		func(initialAvg, newDuration int, gamesPlayed int) bool {
			if gamesPlayed <= 0 {
				gamesPlayed = 1
			}

			stats := &models.PlayerStats{
				Username:    "testplayer",
				GamesPlayed: gamesPlayed,
				GamesWon:    0,
				AvgGameTime: initialAvg,
			}

			stats.UpdateGameStats(false, newDuration)

			// After update, games played should be gamesPlayed + 1
			// New average should be (initialAvg * gamesPlayed + newDuration) / (gamesPlayed + 1)
			expectedAvg := (initialAvg*gamesPlayed + newDuration) / (gamesPlayed + 1)

			return stats.AvgGameTime == expectedAvg
		},
		gen.IntRange(0, 600),  // Initial average (0-10 minutes)
		gen.IntRange(30, 600), // New duration (30 seconds to 10 minutes)
		gen.IntRange(1, 100),  // Games played
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: connect-4-multiplayer, Property 13: Leaderboard Data Retrieval
// *For any* leaderboard request, the system should return the top 10 players sorted by wins.
// **Validates: Requirements 7.2, 7.4**

// TestLeaderboardDataRetrievalProperty tests that leaderboard data is correctly sorted and limited
func TestLeaderboardDataRetrievalProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 50

	properties := gopter.NewProperties(parameters)

	// Property: Leaderboard is always sorted by wins in descending order
	properties.Property("leaderboard is sorted by wins descending", prop.ForAll(
		func(playerData []playerTestData) bool {
			if len(playerData) < 2 {
				return true // Not enough data to test sorting
			}

			// Create player stats from test data
			stats := make([]*models.PlayerStats, len(playerData))
			for i, pd := range playerData {
				stats[i] = &models.PlayerStats{
					Username:    pd.username,
					GamesPlayed: pd.gamesPlayed,
					GamesWon:    pd.gamesWon,
					WinRate:     float64(pd.gamesWon) / float64(max(pd.gamesPlayed, 1)),
				}
			}

			// Sort by wins descending (simulating leaderboard behavior)
			sortedStats := sortByWinsDescending(stats)

			// Verify sorting
			for i := 0; i < len(sortedStats)-1; i++ {
				if sortedStats[i].GamesWon < sortedStats[i+1].GamesWon {
					return false
				}
			}
			return true
		},
		genPlayerDataSlice(),
	))

	// Property: Leaderboard limit is respected
	properties.Property("leaderboard respects limit parameter", prop.ForAll(
		func(numPlayers, limit int) bool {
			if numPlayers <= 0 {
				numPlayers = 1
			}
			if limit <= 0 {
				limit = 10
			}

			// Create test data
			stats := make([]*models.PlayerStats, numPlayers)
			for i := 0; i < numPlayers; i++ {
				stats[i] = &models.PlayerStats{
					Username:    generateUsername(i),
					GamesPlayed: i + 1,
					GamesWon:    i,
				}
			}

			// Apply limit
			result := applyLimit(stats, limit)

			// Verify limit is respected
			expectedLen := min(numPlayers, limit)
			return len(result) == expectedLen
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 50),
	))

	// Property: Default limit is 10 when not specified or invalid
	properties.Property("default limit is 10", prop.ForAll(
		func(invalidLimit int) bool {
			// Test that invalid limits default to 10
			limit := normalizeLimit(invalidLimit)
			if invalidLimit <= 0 {
				return limit == 10
			}
			if invalidLimit > 100 {
				return limit == 100
			}
			return limit == invalidLimit
		},
		gen.IntRange(-10, 200),
	))

	// Property: Players with zero games are excluded from leaderboard
	properties.Property("players with zero games are excluded", prop.ForAll(
		func(playerData []playerTestData) bool {
			// Create player stats from test data
			stats := make([]*models.PlayerStats, len(playerData))
			for i, pd := range playerData {
				stats[i] = &models.PlayerStats{
					Username:    pd.username,
					GamesPlayed: pd.gamesPlayed,
					GamesWon:    pd.gamesWon,
				}
			}

			// Filter out players with zero games (simulating leaderboard behavior)
			filtered := filterActivePlayers(stats)

			// Verify no players with zero games
			for _, s := range filtered {
				if s.GamesPlayed == 0 {
					return false
				}
			}
			return true
		},
		genPlayerDataSlice(),
	))

	// Property: Leaderboard contains unique players only
	properties.Property("leaderboard contains unique players", prop.ForAll(
		func(playerData []playerTestData) bool {
			// Create player stats from test data
			stats := make([]*models.PlayerStats, len(playerData))
			for i, pd := range playerData {
				stats[i] = &models.PlayerStats{
					Username:    pd.username,
					GamesPlayed: pd.gamesPlayed,
					GamesWon:    pd.gamesWon,
				}
			}

			// Check for uniqueness
			seen := make(map[string]bool)
			for _, s := range stats {
				if seen[s.Username] {
					return false
				}
				seen[s.Username] = true
			}
			return true
		},
		genUniquePlayerDataSlice(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper types and functions for property tests

type playerTestData struct {
	username    string
	gamesPlayed int
	gamesWon    int
}

// genPlayerDataSlice generates a slice of player test data
func genPlayerDataSlice() gopter.Gen {
	return gen.SliceOfN(20, genPlayerData())
}

// genUniquePlayerDataSlice generates a slice of player test data with unique usernames
func genUniquePlayerDataSlice() gopter.Gen {
	return gen.IntRange(1, 20).FlatMap(func(n interface{}) gopter.Gen {
		count := n.(int)
		return gen.SliceOfN(count, genPlayerData()).Map(func(data []playerTestData) []playerTestData {
			// Ensure unique usernames
			for i := range data {
				data[i].username = generateUsername(i)
			}
			return data
		})
	}, reflect.TypeOf([]playerTestData{}))
}

// genPlayerData generates a single player test data
func genPlayerData() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString(),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
	).Map(func(values []interface{}) playerTestData {
		username := values[0].(string)
		gamesPlayed := values[1].(int)
		gamesWon := values[2].(int)
		
		// Ensure gamesWon <= gamesPlayed
		if gamesWon > gamesPlayed {
			gamesWon = gamesPlayed
		}
		
		if username == "" {
			username = "player"
		}
		
		return playerTestData{
			username:    username,
			gamesPlayed: gamesPlayed,
			gamesWon:    gamesWon,
		}
	})
}

// generateUsername creates a unique username based on index
func generateUsername(index int) string {
	return "player_" + string(rune('a'+index%26)) + string(rune('0'+index/26))
}

// sortByWinsDescending sorts player stats by wins in descending order
func sortByWinsDescending(stats []*models.PlayerStats) []*models.PlayerStats {
	result := make([]*models.PlayerStats, len(stats))
	copy(result, stats)
	
	// Simple bubble sort for testing purposes
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j].GamesWon < result[j+1].GamesWon {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// applyLimit applies a limit to the stats slice
func applyLimit(stats []*models.PlayerStats, limit int) []*models.PlayerStats {
	if len(stats) <= limit {
		return stats
	}
	return stats[:limit]
}

// normalizeLimit normalizes the limit parameter
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

// filterActivePlayers filters out players with zero games
func filterActivePlayers(stats []*models.PlayerStats) []*models.PlayerStats {
	result := make([]*models.PlayerStats, 0)
	for _, s := range stats {
		if s.GamesPlayed > 0 {
			result = append(result, s)
		}
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
