//go:build property
// +build property

package matchmaking_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/mock"

	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/pkg/models"
)

// MockGameService is defined in mocks_test.go

// Feature: connect-4-multiplayer, Property 1: Matchmaking Queue Management
func TestMatchmakingQueueManagementProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1: Matchmaking Queue Management
	// For any player with a valid username, joining the matchmaking queue should either 
	// pair them with another waiting player or start a bot game within 10 seconds, 
	// ensuring no duplicate usernames in active sessions.
	// Validates: Requirements 1.1, 1.2, 1.3, 1.5
	properties.Property("matchmaking queue management", prop.ForAll(
		func(usernames []string) bool {
			// Filter to valid, unique usernames (3-20 chars, no duplicates)
			validUsernames := make([]string, 0)
			seen := make(map[string]bool)
			for _, username := range usernames {
				if len(username) >= 3 && len(username) <= 20 && !seen[username] {
					validUsernames = append(validUsernames, username)
					seen[username] = true
				}
			}

			// Skip if no valid usernames
			if len(validUsernames) == 0 {
				return true
			}

			// Create mock game service
			mockGameService := new(MockGameService)
			
			// Mock: no players in active games initially
			for _, username := range validUsernames {
				mockGameService.On("GetActiveSessionByPlayer", mock.Anything, username).Return(nil, fmt.Errorf("not found"))
			}
			
			// Mock: game creation for pairs and bot games
			mockGameService.On("CreateSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
				&models.GameSession{
					ID:      "game-1",
					Player1: "player1",
					Player2: "player2",
					Status:  models.StatusInProgress,
				},
				nil,
			).Maybe() // Allow multiple calls with different parameters

			// Create matchmaking service with short timeout for testing
			config := &matchmaking.ServiceConfig{
				MatchTimeout:  1 * time.Second, // Short timeout for property test
				MatchInterval: 50 * time.Millisecond,
				Logger:        slog.Default(),
			}
			service := matchmaking.NewMatchmakingService(mockGameService, config)

			// Track game creations
			var createdGames []*models.GameSession
			var botGames []*models.GameSession
			
			service.SetGameCreatedCallback(func(ctx context.Context, player1, player2 string, gameSession *models.GameSession) error {
				createdGames = append(createdGames, gameSession)
				return nil
			})
			
			service.SetBotGameCallback(func(ctx context.Context, player string, gameSession *models.GameSession) error {
				botGames = append(botGames, gameSession)
				return nil
			})

			// Start matchmaking
			ctx := context.Background()
			if err := service.StartMatchmaking(ctx); err != nil {
				return false
			}
			defer service.StopMatchmaking()

			// Add all players to queue
			for _, username := range validUsernames {
				if _, err := service.JoinQueue(ctx, username); err != nil {
					// Should not fail for valid usernames (Requirement 1.1)
					return false
				}
			}

			// Wait for matchmaking to process (longer than timeout)
			time.Sleep(1500 * time.Millisecond)

			// Verify all players were processed
			finalQueueLength := service.GetQueueLength(ctx)
			if finalQueueLength != 0 {
				// All players should be removed from queue (either matched or timed out)
				return false
			}

			// Calculate expected outcomes
			expectedPairs := len(validUsernames) / 2
			expectedBotGames := len(validUsernames) % 2

			// Verify game creation counts
			if len(createdGames) != expectedPairs {
				// Should create games for pairs (Requirement 1.2)
				return false
			}

			if len(botGames) != expectedBotGames {
				// Should create bot games for remaining players (Requirement 1.3)
				return false
			}

			// Verify no duplicate usernames in games (Requirement 1.5)
			allPlayers := make(map[string]bool)
			for _, game := range createdGames {
				if allPlayers[game.Player1] || allPlayers[game.Player2] {
					return false // Duplicate player found
				}
				allPlayers[game.Player1] = true
				allPlayers[game.Player2] = true
			}
			for _, game := range botGames {
				if allPlayers[game.Player1] {
					return false // Duplicate player found
				}
				allPlayers[game.Player1] = true
			}

			// Verify all players were assigned to games
			if len(allPlayers) != len(validUsernames) {
				return false
			}

			return true
		},
		genUsernameList(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genUsernameList generates a list of usernames for testing
func genUsernameList() gopter.Gen {
	return gen.SliceOf(genUsername()).SuchThat(func(usernames []string) bool {
		// Limit to reasonable size for testing
		return len(usernames) <= 10
	})
}

// genUsername generates valid and invalid usernames for testing
func genUsername() gopter.Gen {
	return gen.Frequency(map[int]gopter.Gen{
		// Valid usernames (3-20 chars) - most common
		8: gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
		// Some invalid usernames for edge case testing
		1: gen.Const(""), // Empty username
		2: gen.Const("ab"), // Too short
	})
}

// Feature: connect-4-multiplayer, Property 1a: Queue Timeout Behavior
func TestMatchmakingTimeoutProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1a: Single player timeout behavior
	// For any single player in the queue, they should get a bot game within the timeout period
	properties.Property("single player gets bot game within timeout", prop.ForAll(
		func(username string) bool {
			// Skip invalid usernames
			if len(username) < 3 || len(username) > 20 {
				return true
			}

			// Create mock game service
			mockGameService := new(MockGameService)
			mockGameService.On("GetActiveSessionByPlayer", mock.Anything, username).Return(nil, fmt.Errorf("not found"))
			mockGameService.On("CreateSession", mock.Anything, username, mock.AnythingOfType("string")).Return(
				&models.GameSession{
					ID:      "bot-game-1",
					Player1: username,
					Player2: "bot_123",
					Status:  models.StatusInProgress,
				},
				nil,
			)

			// Create service with 500ms timeout
			config := &matchmaking.ServiceConfig{
				MatchTimeout:  500 * time.Millisecond,
				MatchInterval: 50 * time.Millisecond,
				Logger:        slog.Default(),
			}
			service := matchmaking.NewMatchmakingService(mockGameService, config)

			// Track bot game creation
			var botGameCreated bool
			service.SetBotGameCallback(func(ctx context.Context, player string, gameSession *models.GameSession) error {
				botGameCreated = true
				return nil
			})

			// Start matchmaking
			ctx := context.Background()
			if err := service.StartMatchmaking(ctx); err != nil {
				return false
			}
			defer service.StopMatchmaking()

			// Add player to queue
			if _, err := service.JoinQueue(ctx, username); err != nil {
				return false
			}

			// Wait for timeout + processing time
			time.Sleep(700 * time.Millisecond)

			// Verify bot game was created
			if !botGameCreated {
				return false
			}

			// Verify queue is empty
			if service.GetQueueLength(ctx) != 0 {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: connect-4-multiplayer, Property 1b: Duplicate Username Prevention
func TestMatchmakingDuplicatePreventionProperty(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1b: Duplicate username prevention
	// For any username already in an active game, joining the queue should be rejected
	properties.Property("duplicate usernames in active games are rejected", prop.ForAll(
		func(username string) bool {
			// Skip invalid usernames
			if len(username) < 3 || len(username) > 20 {
				return true
			}

			// Create mock game service
			mockGameService := new(MockGameService)
			
			// Mock: player already in active game
			activeGame := &models.GameSession{
				ID:      "existing-game",
				Player1: username,
				Player2: "other-player",
				Status:  models.StatusInProgress,
			}
			mockGameService.On("GetActiveSessionByPlayer", mock.Anything, username).Return(activeGame, nil)

			// Create service
			service := matchmaking.NewMatchmakingService(mockGameService, matchmaking.DefaultServiceConfig())

			// Try to join queue
			ctx := context.Background()
			_, err := service.JoinQueue(ctx, username)

			// Should be rejected (Requirement 1.5)
			return err != nil
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			return len(s) >= 3 && len(s) <= 20
		}),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}