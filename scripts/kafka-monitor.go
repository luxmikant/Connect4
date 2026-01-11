//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/pkg/models"

	"github.com/gin-gonic/gin"
)

type KafkaMonitor struct {
	producer *analytics.Producer
	messages []models.GameEvent
}

func main() {
	fmt.Println("🖥️  Starting Kafka Monitor Web Interface...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create producer for sending test messages
	producer := analytics.NewProducer(cfg.Kafka)

	monitor := &KafkaMonitor{
		producer: producer,
		messages: make([]models.GameEvent, 0),
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Serve static HTML
	r.GET("/", monitor.serveHTML)
	r.GET("/api/messages", monitor.getMessages)
	r.POST("/api/send-test", monitor.sendTestMessage)

	fmt.Println("🌐 Kafka Monitor running at: http://localhost:8081")
	fmt.Println("📊 Open your browser to monitor Kafka messages")
	fmt.Println("🔧 Use the interface to send test messages to Confluent Cloud")

	log.Fatal(http.ListenAndServe(":8081", r))
}

func (m *KafkaMonitor) serveHTML(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Kafka Cloud Monitor</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .controls { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .messages { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        button { background: #3498db; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; margin: 5px; }
        button:hover { background: #2980b9; }
        .message { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 4px; background: #f9f9f9; }
        .message-header { font-weight: bold; color: #2c3e50; margin-bottom: 10px; }
        .message-body { font-family: monospace; background: #ecf0f1; padding: 10px; border-radius: 4px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .info { background: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 Kafka Cloud Monitor</h1>
            <p>Monitor and test your Confluent Cloud Kafka connection</p>
        </div>
        
        <div class="controls">
            <h3>📤 Send Test Messages</h3>
            <button onclick="sendTestMessage('player_joined')">Send Player Joined</button>
            <button onclick="sendTestMessage('game_started')">Send Game Started</button>
            <button onclick="sendTestMessage('move_made')">Send Move Made</button>
            <button onclick="sendTestMessage('game_completed')">Send Game Completed</button>
            <button onclick="clearMessages()">Clear Messages</button>
            <div id="status"></div>
        </div>
        
        <div class="messages">
            <h3>📨 Sent Messages</h3>
            <div id="messageList">
                <div class="info">No messages sent yet. Click the buttons above to send test messages to Kafka.</div>
            </div>
        </div>
    </div>

    <script>
        let messageCount = 0;
        
        function sendTestMessage(type) {
            fetch('/api/send-test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: type })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showStatus('✅ Message sent successfully!', 'success');
                    addMessage(data.event);
                } else {
                    showStatus('❌ Failed to send message: ' + data.error, 'error');
                }
            })
            .catch(error => {
                showStatus('❌ Error: ' + error, 'error');
            });
        }
        
        function addMessage(event) {
            messageCount++;
            const messageList = document.getElementById('messageList');
            if (messageCount === 1) {
                messageList.innerHTML = '';
            }
            
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message';
            messageDiv.innerHTML = 
                '<div class="message-header">' + 
                '#' + messageCount + ' - ' + event.event_type + ' (' + new Date(event.timestamp).toLocaleTimeString() + ')' +
                '</div>' +
                '<div class="message-body">' + JSON.stringify(event, null, 2) + '</div>';
            
            messageList.insertBefore(messageDiv, messageList.firstChild);
        }
        
        function showStatus(message, type) {
            const status = document.getElementById('status');
            status.innerHTML = '<div class="status ' + type + '">' + message + '</div>';
            setTimeout(() => { status.innerHTML = ''; }, 5000);
        }
        
        function clearMessages() {
            document.getElementById('messageList').innerHTML = 
                '<div class="info">Messages cleared. Send new test messages above.</div>';
            messageCount = 0;
        }
    </script>
</body>
</html>`
	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

func (m *KafkaMonitor) getMessages(c *gin.Context) {
	c.JSON(200, gin.H{"messages": m.messages})
}

func (m *KafkaMonitor) sendTestMessage(c *gin.Context) {
	var request struct {
		Type string `json:"type"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := context.Background()
	var err error
	var event *models.GameEvent

	gameID := fmt.Sprintf("monitor-test-%d", time.Now().Unix())

	switch request.Type {
	case "player_joined":
		err = m.producer.SendPlayerJoined(ctx, gameID, "monitor-player")
		event = &models.GameEvent{
			EventType: models.EventPlayerJoined,
			GameID:    gameID,
			PlayerID:  "monitor-player",
			Timestamp: time.Now(),
		}
	case "game_started":
		err = m.producer.SendGameStarted(ctx, gameID, "player1", "player2")
		event = &models.GameEvent{
			EventType: models.EventGameStarted,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now(),
		}
	case "move_made":
		err = m.producer.SendMoveMade(ctx, gameID, "player1", 3, 0, 1)
		event = &models.GameEvent{
			EventType: models.EventMoveMade,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now(),
		}
	case "game_completed":
		err = m.producer.SendGameCompleted(ctx, gameID, "player1", "player2", 5*time.Minute)
		event = &models.GameEvent{
			EventType: models.EventGameCompleted,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now(),
		}
	default:
		c.JSON(400, gin.H{"success": false, "error": "Unknown message type"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Store message for display
	m.messages = append(m.messages, *event)

	c.JSON(200, gin.H{"success": true, "event": event})
}
