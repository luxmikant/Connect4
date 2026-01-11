package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/config"
)

func main() {
	fmt.Println("🔍 Testing Kafka Cloud Connection (Confluent Cloud)")
	fmt.Println(strings.Repeat("=", 60))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Display configuration (without secrets)
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Bootstrap Servers: %s\n", cfg.Kafka.BootstrapServers)
	fmt.Printf("   Topic: %s\n", cfg.Kafka.Topic)
	fmt.Printf("   Consumer Group: %s\n", cfg.Kafka.ConsumerGroup)
	fmt.Printf("   API Key: %s...\n", cfg.Kafka.APIKey[:8])
	fmt.Printf("   API Secret: %s...\n", cfg.Kafka.APISecret[:8])
	fmt.Println()

	// Test 1: Create Producer
	fmt.Println("🔧 Test 1: Creating Kafka Producer...")
	producer := analytics.NewProducer(cfg.Kafka)
	if producer == nil {
		log.Fatalf("❌ Failed to create producer")
	}
	fmt.Println("✅ Producer created successfully")

	// Test 2: Send Test Events
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n📤 Test 2: Sending Test Events...")

	// Send different types of events
	testEvents := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Player Joined Event",
			fn: func() error {
				return producer.SendPlayerJoined(ctx, "test-game-123", "test-player-1")
			},
		},
		{
			name: "Game Started Event",
			fn: func() error {
				return producer.SendGameStarted(ctx, "test-game-123", "player1", "player2")
			},
		},
		{
			name: "Move Event",
			fn: func() error {
				return producer.SendMoveMade(ctx, "test-game-123", "player1", 3, 0, 1)
			},
		},
		{
			name: "Game Completed Event",
			fn: func() error {
				return producer.SendGameCompleted(ctx, "test-game-123", "player1", "player2", 5*time.Minute)
			},
		},
	}

	successCount := 0
	for i, test := range testEvents {
		fmt.Printf("   %d. Sending %s... ", i+1, test.name)

		start := time.Now()
		err := test.fn()
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v)\n", err)
		} else {
			fmt.Printf("✅ SUCCESS (%v)\n", duration)
			successCount++
		}
	}

	// Test 3: Performance Test
	fmt.Println("\n⚡ Test 3: Performance Test (10 rapid events)...")
	start := time.Now()

	for i := 0; i < 10; i++ {
		err := producer.SendPlayerJoined(ctx, fmt.Sprintf("perf-test-%d", i), "perf-player")
		if err != nil {
			fmt.Printf("❌ Event %d failed: %v\n", i+1, err)
		}
	}

	totalDuration := time.Since(start)
	avgDuration := totalDuration / 10

	fmt.Printf("✅ Sent 10 events in %v (avg: %v per event)\n", totalDuration, avgDuration)

	// Close producer
	producer.Close()

	// Summary
	fmt.Println("\n📊 Test Summary:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ Producer Creation: SUCCESS\n")
	fmt.Printf("✅ Events Sent: %d/%d\n", successCount, len(testEvents))
	fmt.Printf("✅ Performance: %v avg per event\n", avgDuration)

	if successCount == len(testEvents) {
		fmt.Println("\n🎉 ALL TESTS PASSED - Kafka Cloud is working perfectly!")
		fmt.Println("Your Confluent Cloud connection is ready for production.")
	} else {
		fmt.Printf("\n⚠️  %d/%d tests failed - Check your Kafka configuration\n", len(testEvents)-successCount, len(testEvents))
	}

	fmt.Println("\n💡 Next Steps:")
	fmt.Println("   1. Check Confluent Cloud console for received messages")
	fmt.Println("   2. Start analytics service: go run cmd/analytics/main.go")
	fmt.Println("   3. Monitor message consumption in real-time")
}
