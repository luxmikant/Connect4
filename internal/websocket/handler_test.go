package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"connect4-multiplayer/pkg/models"
)

func TestMessageCreation(t *testing.T) {
	t.Run("CreateJoinGameMessage", func(t *testing.T) {
		msg := CreateJoinGameMessage("testuser", "bot")
		assert.Equal(t, MessageTypeJoinGame, msg.Type)
		assert.Equal(t, "testuser", msg.Payload["username"])
		assert.Equal(t, "bot", msg.Payload["gameType"])
	})

	t.Run("CreateMakeMoveMessage", func(t *testing.T) {
		msg := CreateMakeMoveMessage("game123", 3)
		assert.Equal(t, MessageTypeMakeMove, msg.Type)
		assert.Equal(t, "game123", msg.Payload["gameId"])
		assert.Equal(t, 3, msg.Payload["column"])
	})

	t.Run("CreateReconnectMessage", func(t *testing.T) {
		msg := CreateReconnectMessage("game123", "testuser")
		assert.Equal(t, MessageTypeReconnect, msg.Type)
		assert.Equal(t, "game123", msg.Payload["gameId"])
		assert.Equal(t, "testuser", msg.Payload["username"])
	})

	t.Run("CreateGameStartedMessage", func(t *testing.T) {
		msg := CreateGameStartedMessage("game123", "opponent", "red", "red", true)
		assert.Equal(t, MessageTypeGameStarted, msg.Type)
		assert.Equal(t, "game123", msg.Payload["gameId"])
		assert.Equal(t, "opponent", msg.Payload["opponent"])
		assert.Equal(t, "red", msg.Payload["yourColor"])
		assert.Equal(t, "red", msg.Payload["currentTurn"])
		assert.Equal(t, true, msg.Payload["isBot"])
	})

	t.Run("CreateMoveMadeMessage", func(t *testing.T) {
		board := models.NewBoard()
		msg := CreateMoveMadeMessage("game123", "player1", 3, 0, board, "yellow", 1)
		assert.Equal(t, MessageTypeMoveMade, msg.Type)
		assert.Equal(t, "game123", msg.Payload["gameId"])
		assert.Equal(t, "player1", msg.Payload["player"])
		assert.Equal(t, 3, msg.Payload["column"])
		assert.Equal(t, 0, msg.Payload["row"])
		assert.Equal(t, "yellow", msg.Payload["nextTurn"])
		assert.Equal(t, 1, msg.Payload["moveCount"])
	})

	t.Run("CreateGameEndedMessage", func(t *testing.T) {
		winner := "player1"
		msg := CreateGameEndedMessage("game123", &winner, "connect_four", 120)
		assert.Equal(t, MessageTypeGameEnded, msg.Type)
		assert.Equal(t, "game123", msg.Payload["gameId"])
		assert.Equal(t, &winner, msg.Payload["winner"])
		assert.Equal(t, "connect_four", msg.Payload["reason"])
		assert.Equal(t, 120, msg.Payload["duration"])
	})

	t.Run("CreateErrorMessage", func(t *testing.T) {
		msg := CreateErrorMessage("INVALID_MOVE", "Invalid move", "Column is full")
		assert.Equal(t, MessageTypeError, msg.Type)
		assert.Equal(t, "INVALID_MOVE", msg.Payload["code"])
		assert.Equal(t, "Invalid move", msg.Payload["message"])
		assert.Equal(t, "Column is full", msg.Payload["details"])
	})

	t.Run("CreatePongMessage", func(t *testing.T) {
		msg := CreatePongMessage()
		assert.Equal(t, MessageTypePong, msg.Type)
		assert.NotNil(t, msg.Payload)
	})
}

func TestMessageSerialization(t *testing.T) {
	t.Run("ToJSON and FromJSON", func(t *testing.T) {
		originalMsg := CreateJoinGameMessage("testuser", "pvp")
		
		// Serialize to JSON
		jsonData, err := originalMsg.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Deserialize from JSON
		parsedMsg, err := FromJSON(jsonData)
		require.NoError(t, err)
		
		assert.Equal(t, originalMsg.Type, parsedMsg.Type)
		assert.Equal(t, originalMsg.Payload["username"], parsedMsg.Payload["username"])
		assert.Equal(t, originalMsg.Payload["gameType"], parsedMsg.Payload["gameType"])
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)
		_, err := FromJSON(invalidJSON)
		assert.Error(t, err)
	})
}

func TestConnectionConfig(t *testing.T) {
	t.Run("DefaultConnectionConfig", func(t *testing.T) {
		config := DefaultConnectionConfig()
		
		assert.Equal(t, 10*time.Second, config.WriteWait)
		assert.Equal(t, 60*time.Second, config.PongWait)
		assert.Equal(t, 54*time.Second, config.PingPeriod)
		assert.Equal(t, int64(512), config.MaxMessageSize)
		
		// Verify ping period is less than pong wait (required for proper operation)
		assert.True(t, config.PingPeriod < config.PongWait)
	})
}

func TestWebSocketErrors(t *testing.T) {
	t.Run("Error definitions", func(t *testing.T) {
		assert.NotEmpty(t, ErrUserNotConnected.Error())
		assert.NotEmpty(t, ErrGameNotFound.Error())
		assert.NotEmpty(t, ErrInvalidMessage.Error())
		assert.NotEmpty(t, ErrUnauthorized.Error())
		assert.NotEmpty(t, ErrConnectionClosed.Error())
		assert.NotEmpty(t, ErrMessageTooLarge.Error())
		assert.NotEmpty(t, ErrRateLimitExceeded.Error())
		assert.NotEmpty(t, ErrInvalidGameState.Error())
		assert.NotEmpty(t, ErrPlayerNotInGame.Error())
		assert.NotEmpty(t, ErrGameAlreadyEnded.Error())
	})
}