---
inclusion: always
---

# Development Guidelines

## Code Quality Standards

### Error Handling
- Always use structured error handling with proper context
- Return errors from all functions that can fail
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Log errors at appropriate levels (Error, Warn, Info, Debug)
- Never ignore errors - handle or explicitly document why ignored

```go
if err := gameService.CreateGame(ctx, req); err != nil {
    log.Error("failed to create game", "error", err, "playerID", req.PlayerID)
    return fmt.Errorf("creating game: %w", err)
}
```

### Context Usage
- Pass `context.Context` as first parameter to all service methods
- Use context for cancellation, timeouts, and request-scoped values
- Set appropriate timeouts for external calls (database, Kafka)
- Propagate context through the entire call chain

### Logging Standards
- Use structured logging with key-value pairs
- Include relevant context (playerID, gameID, sessionID)
- Log at entry/exit points of critical operations
- Use consistent field names across the application

```go
log.Info("game created", 
    "gameID", game.ID, 
    "playerID", req.PlayerID, 
    "gameType", req.GameType,
    "duration", time.Since(start))
```

## Testing Requirements

### Unit Testing
- Achieve minimum 80% code coverage for business logic
- Use table-driven tests for multiple scenarios
- Mock all external dependencies (database, Kafka, Redis)
- Test both success and failure paths
- Include edge cases and boundary conditions

### Property-Based Testing
- Required for all game logic functions (move validation, win detection)
- Use gopter generators for game states and moves
- Test invariants that should always hold true
- Include shrinking to find minimal failing cases

```go
//go:build property

func TestGameLogicProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("valid moves never cause invalid game state", prop.ForAll(
        func(board Board, move Move) bool {
            if !move.IsValid(board) {
                return true // Skip invalid moves
            }
            newBoard := board.ApplyMove(move)
            return newBoard.IsValid()
        },
        genBoard(), genMove(),
    ))
}
```

### Integration Testing
- Test WebSocket connection lifecycle
- Verify database transactions and rollbacks
- Test Kafka message production and consumption
- Include timeout and reconnection scenarios

## Performance Guidelines

### Database Operations
- Use GORM preloading to prevent N+1 queries
- Implement database connection pooling
- Add database query timeouts (5 seconds max)
- Use transactions for multi-table operations
- Index frequently queried columns

```go
// Good: Preload related data
var games []models.Game
db.Preload("Players").Preload("Moves").Find(&games)

// Bad: N+1 query problem
for _, game := range games {
    db.Find(&game.Players, "game_id = ?", game.ID)
}
```

### WebSocket Management
- Implement connection pooling with cleanup
- Set read/write deadlines on connections
- Use buffered channels for message queuing
- Implement graceful shutdown with connection draining
- Monitor connection count and memory usage

### Memory Management
- Clean up completed game sessions within 5 minutes
- Use object pooling for frequently allocated structs
- Implement proper goroutine lifecycle management
- Monitor memory usage and implement alerts

## Real-Time Requirements

### WebSocket Message Handling
- Process messages asynchronously to prevent blocking
- Implement message queuing with backpressure
- Use separate goroutines for read/write operations
- Handle connection drops gracefully with reconnection logic

```go
func (h *GameHandler) handleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // Set timeouts
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
    
    // Handle messages in separate goroutines
    go h.handleReads(conn)
    go h.handleWrites(conn)
}
```

### Game State Synchronization
- Broadcast state changes to all connected players within 100ms
- Use optimistic updates on client with server reconciliation
- Implement conflict resolution for simultaneous moves
- Maintain game state consistency across reconnections

## Bot AI Implementation

### Minimax Algorithm
- Implement iterative deepening for time management
- Use transposition tables for position caching
- Apply move ordering heuristics (center columns first)
- Ensure deterministic behavior for testing

```go
func (b *Bot) GetBestMove(board Board, depth int, timeLimit time.Duration) Move {
    start := time.Now()
    bestMove := Move{}
    
    for d := 1; d <= depth; d++ {
        if time.Since(start) > timeLimit {
            break
        }
        move := b.minimax(board, d, -math.Inf(1), math.Inf(1), true)
        if move.IsValid(board) {
            bestMove = move
        }
    }
    
    return bestMove
}
```

### Performance Constraints
- Bot moves must complete within 800ms (leaving 200ms buffer)
- Implement timeout handling with best available move
- Use alpha-beta pruning for search optimization
- Cache position evaluations between moves

## Analytics Implementation

### Event Tracking
- Send analytics events asynchronously to prevent blocking gameplay
- Use structured event schemas with versioning
- Implement event batching for performance
- Include correlation IDs for event tracing

```go
type GameEvent struct {
    EventType   string    `json:"eventType"`
    GameID      string    `json:"gameId"`
    PlayerID    string    `json:"playerId"`
    Timestamp   time.Time `json:"timestamp"`
    Properties  map[string]interface{} `json:"properties"`
    Version     string    `json:"version"`
}
```

### Kafka Integration
- Use at-least-once delivery semantics
- Implement producer retry logic with exponential backoff
- Set appropriate batch size and linger time
- Monitor producer metrics and lag

## Security Practices

### Input Validation
- Validate all user inputs at API boundaries
- Sanitize data before database operations
- Use parameterized queries to prevent SQL injection
- Implement rate limiting on API endpoints

### WebSocket Security
- Validate origin headers for WebSocket connections
- Implement authentication token validation
- Use HTTPS/WSS in production
- Limit message size and frequency per connection

## Monitoring and Observability

### Metrics Collection
- Track game creation/completion rates
- Monitor WebSocket connection counts
- Measure bot response times
- Track database query performance

### Health Checks
- Implement health check endpoints for all services
- Check database connectivity
- Verify Kafka producer/consumer health
- Monitor memory and CPU usage

### Alerting
- Set up alerts for high error rates
- Monitor WebSocket connection failures
- Track bot timeout occurrences
- Alert on database connection pool exhaustion

## Development Workflow

### Code Review Requirements
- All code must pass automated tests
- Require approval from at least one reviewer
- Run linting and formatting checks
- Verify Swagger documentation is updated

### Deployment Process
- Use feature flags for gradual rollouts
- Implement blue-green deployment strategy
- Run smoke tests after deployment
- Monitor key metrics during rollout

### Local Development
- Use Docker Compose for consistent environment
- Include sample data for testing
- Provide clear setup instructions
- Mock external services for offline development