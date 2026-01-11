//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"
)

// TestMain sets up the test environment for integration tests
func TestMain(m *testing.M) {
	// Set up test environment variables if not already set
	setupTestEnvironment()
	
	// Run tests
	code := m.Run()
	
	// Clean up
	cleanupTestEnvironment()
	
	os.Exit(code)
}

// setupTestEnvironment configures the test environment
func setupTestEnvironment() {
	// Set test environment
	if os.Getenv("ENVIRONMENT") == "" {
		os.Setenv("ENVIRONMENT", "test")
	}
	
	// Set test database URL if not provided
	if os.Getenv("TEST_DATABASE_URL") == "" && os.Getenv("DATABASE_URL") != "" {
		// Use main database URL for tests if no separate test database is configured
		// In production, you should use a separate test database
		os.Setenv("TEST_DATABASE_URL", os.Getenv("DATABASE_URL"))
	}
	
	// Set test Kafka topic prefix
	if os.Getenv("KAFKA_TOPIC") == "" {
		os.Setenv("KAFKA_TOPIC", "test-game-events")
	}
	
	// Set test consumer group
	if os.Getenv("KAFKA_CONSUMER_GROUP") == "" {
		os.Setenv("KAFKA_CONSUMER_GROUP", "test-analytics-service")
	}
	
	// Set server port for tests
	if os.Getenv("SERVER_PORT") == "" {
		os.Setenv("SERVER_PORT", "0") // Use random port for tests
	}
}

// cleanupTestEnvironment cleans up after tests
func cleanupTestEnvironment() {
	// Any cleanup needed after all tests complete
}