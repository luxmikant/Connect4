# Integration Testing Guide

This document provides comprehensive guidance for running integration tests for the Connect 4 Multiplayer System with cloud services.

## Overview

The integration test suite validates the complete system functionality including:

- **End-to-End Game Flow**: Complete game sessions from matchmaking to completion
- **Cloud Service Integration**: Supabase PostgreSQL and Confluent Cloud Kafka
- **Performance Testing**: Load testing with multiple concurrent connections
- **WebSocket Communication**: Real-time message handling and reconnection
- **Bot AI Performance**: Response time validation under load
- **Database Performance**: Connection pooling and query optimization

## Test Structure

```
tests/integration/
├── e2e_test.go           # End-to-end integration tests
├── performance_test.go   # Performance and load tests
└── config_test.go        # Test configuration and setup
```

## Prerequisites

### Environment Setup

1. **Cloud Services Configuration**
   - Supabase PostgreSQL database with connection string
   - Confluent Cloud Kafka cluster with API credentials
   - Redis Cloud instance (optional)

2. **Environment Variables**
   ```bash
   # Required
   DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres?sslmode=require
   
   # Optional (for Kafka tests)
   KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.region.provider.confluent.cloud:9092
   KAFKA_API_KEY=your-api-key
   KAFKA_API_SECRET=your-api-secret
   
   # Test-specific (optional)
   TEST_DATABASE_URL=postgresql://postgres:password@test-db.supabase.co:5432/postgres?sslmode=require
   ```

3. **Go Dependencies**
   ```bash
   go mod download
   ```

## Running Tests

### Quick Start

```bash
# Run all integration tests
./scripts/run-integration-tests.sh

# Windows PowerShell
.\scripts\run-integration-tests.ps1
```

### Specific Test Suites

```bash
# End-to-end tests only
./scripts/run-integration-tests.sh e2e

# Performance tests only
./scripts/run-integration-tests.sh performance

# Kafka integration tests only
./scripts/run-integration-tests.sh kafka
```

### Manual Test Execution

```bash
# Run with Go test command
go test -tags=integration -v ./tests/integration -run=TestE2ETestSuite

# Run with timeout
go test -tags=integration -timeout=15m -v ./tests/integration

# Run specific test
go test -tags=integration -v ./tests/integration -run=TestCompleteGameFlow
```

## Test Suites

### 1. End-to-End Integration Tests (`TestE2ETestSuite`)

**Duration**: ~10-15 minutes  
**Purpose**: Validates complete system functionality with cloud services

**Test Cases**:
- `TestCompleteGameFlow`: Full game from matchmaking to completion
- `TestBotGameFlow`: Player vs bot game scenarios
- `TestMultiClientScenarios`: Multiple concurrent games
- `TestReconnectionScenarios`: WebSocket reconnection handling
- `TestKafkaIntegration`: Analytics event processing

**Validation Points**:
- ✅ Player creation and authentication
- ✅ WebSocket connection establishment
- ✅ Matchmaking queue management
- ✅ Real-time game state synchronization
- ✅ Bot AI decision making
- ✅ Game completion and statistics
- ✅ Analytics event publishing
- ✅ Leaderboard updates

### 2. Performance Tests (`TestPerformanceTestSuite`)

**Duration**: ~15-20 minutes  
**Purpose**: Validates system performance under load

**Test Cases**:
- `TestConcurrentGameSessions`: 10+ concurrent games
- `TestWebSocketPerformance`: 50+ concurrent connections
- `TestBotResponseTimes`: Bot AI performance under load
- `TestDatabasePerformance`: Database query optimization
- `TestSupabaseConnectionLimits`: Connection pool behavior

**Performance Requirements**:
- ✅ Average response time < 100ms
- ✅ Bot response time < 800ms
- ✅ Database query time < 50ms
- ✅ Success rate > 95%
- ✅ Connection error rate < 5%

### 3. Cloud Service Integration

**Supabase PostgreSQL**:
- Connection pooling validation
- Migration execution
- CRUD operation performance
- Connection limit handling

**Confluent Cloud Kafka**:
- Event publishing reliability
- Consumer group management
- Message serialization/deserialization
- Error handling and retries

## Performance Metrics

The test suite tracks comprehensive performance metrics:

```go
type PerformanceMetrics struct {
    // Connection metrics
    ConnectionCount    int
    MaxConnections     int
    ConnectionErrors   int
    
    // Response time metrics
    ResponseTimes      []time.Duration
    MaxResponseTime    time.Duration
    MinResponseTime    time.Duration
    AvgResponseTime    time.Duration
    
    // Game metrics
    GamesCreated       int
    GamesCompleted     int
    GameErrors         int
    
    // Bot metrics
    BotMoves           int
    BotResponseTimes   []time.Duration
    BotTimeouts        int
    
    // Database metrics
    DatabaseQueries    int
    DatabaseErrors     int
    DatabaseLatency    []time.Duration
}
```

## Test Configuration

### Database Configuration

```go
// Test database setup
db, err := gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
})

// Connection pool optimization
sqlDB.SetMaxOpenConns(50)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(300 * time.Second)
```

### WebSocket Configuration

```go
// WebSocket test client
conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

// Message handling with timeouts
conn.SetReadDeadline(time.Now().Add(5 * time.Second))
conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
```

### Service Configuration

```go
// Optimized for testing
matchmakingConfig := &matchmaking.ServiceConfig{
    MatchTimeout:  1 * time.Second,
    MatchInterval: 50 * time.Millisecond,
}

// Analytics with test topic
analyticsConfig := &analytics.ServiceConfig{
    KafkaConfig: analytics.KafkaConfig{
        Topic:         "test-game-events",
        ConsumerGroup: "test-analytics-service",
    },
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   ```
   Error: failed to connect to database
   ```
   - Verify `DATABASE_URL` is correct
   - Check Supabase project is active
   - Ensure IP is whitelisted in Supabase

2. **Kafka Connection Errors**
   ```
   Error: failed to create Kafka producer
   ```
   - Verify Kafka credentials are correct
   - Check Confluent Cloud cluster is active
   - Ensure API key has proper permissions

3. **WebSocket Connection Failures**
   ```
   Error: websocket: bad handshake
   ```
   - Check server is running
   - Verify WebSocket endpoint is accessible
   - Check for port conflicts

4. **Test Timeouts**
   ```
   Error: test timed out
   ```
   - Increase test timeout values
   - Check cloud service latency
   - Verify network connectivity

### Debug Mode

Enable verbose logging for debugging:

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run with verbose output
go test -tags=integration -v ./tests/integration -run=TestE2ETestSuite
```

### Test Data Cleanup

Tests automatically clean up test data, but manual cleanup may be needed:

```sql
-- Clean test data from database
DELETE FROM game_events WHERE game_id LIKE 'test-%';
DELETE FROM moves WHERE game_id LIKE 'test-%';
DELETE FROM game_sessions WHERE id LIKE 'test-%';
DELETE FROM player_stats WHERE username LIKE 'test_%';
DELETE FROM players WHERE username LIKE 'test_%';
```

## Continuous Integration

### GitHub Actions Integration

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Run Integration Tests
        env:
          DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
          KAFKA_BOOTSTRAP_SERVERS: ${{ secrets.KAFKA_BOOTSTRAP_SERVERS }}
          KAFKA_API_KEY: ${{ secrets.KAFKA_API_KEY }}
          KAFKA_API_SECRET: ${{ secrets.KAFKA_API_SECRET }}
        run: |
          go test -tags=integration -timeout=20m -v ./tests/integration
```

### Test Reporting

The test runner generates comprehensive reports:

- **integration-test-report.md**: Detailed test results and metrics
- **Performance summary**: Response times, error rates, throughput
- **Cloud service validation**: Connection status and performance
- **Recommendations**: Performance optimization suggestions

## Best Practices

### Test Environment

1. **Use Separate Test Database**: Avoid conflicts with development data
2. **Test Topic Isolation**: Use test-specific Kafka topics
3. **Connection Limits**: Configure appropriate connection pools
4. **Cleanup Strategy**: Ensure test data is properly cleaned up

### Performance Testing

1. **Realistic Load**: Test with realistic concurrent user counts
2. **Gradual Ramp-up**: Increase load gradually to identify bottlenecks
3. **Metric Collection**: Track comprehensive performance metrics
4. **Baseline Comparison**: Compare results against previous runs

### Cloud Service Testing

1. **Credential Management**: Use environment variables for credentials
2. **Service Health**: Verify cloud services are operational before testing
3. **Rate Limiting**: Respect cloud service rate limits
4. **Error Handling**: Test failure scenarios and recovery

## Monitoring and Alerts

### Key Metrics to Monitor

- **Response Time**: Average < 100ms, P95 < 500ms
- **Error Rate**: < 5% for all operations
- **Connection Count**: Monitor WebSocket connections
- **Database Performance**: Query time < 50ms average
- **Bot Performance**: Response time < 800ms

### Alert Thresholds

```yaml
alerts:
  - name: High Response Time
    condition: avg_response_time > 200ms
    
  - name: High Error Rate
    condition: error_rate > 10%
    
  - name: Database Slow Queries
    condition: db_query_time > 100ms
    
  - name: Bot Timeout Rate
    condition: bot_timeout_rate > 15%
```

## Conclusion

The integration test suite provides comprehensive validation of the Connect 4 Multiplayer System with cloud services. Regular execution ensures system reliability, performance, and proper cloud service integration.

For questions or issues, refer to the troubleshooting section or check the project documentation.