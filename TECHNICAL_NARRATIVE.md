# What's the Most Interesting Thing I've Built?

## Connect4 Multiplayer – A Real-Time Game System

The most interesting thing I've built is a real-time multiplayer Connect 4 system. What started as a straightforward game quickly became an exercise in distributed systems, real-time communication, AI, and production deployment—all within a single project.

---

## Why It's Interesting

A lot of projects are CRUD apps: store some data, render it, repeat. Connect4 Multiplayer forced me to think about problems I rarely face in typical web development:

- **Two users need to see the same board state instantly.** A REST API's request-response model doesn't cut it.
- **A bot needs to play moves intelligently without blocking the server.** CPU-intensive AI must coexist with concurrent user connections.
- **Players can drop off mid-game.** Network failures are a given, not an edge case.
- **Analytics must flow without slowing the game.** You can't add I/O latency to a player's move processing path.

Solving all four of these challenges in one cohesive system is what made this project genuinely interesting.

---

## Key Technical Decisions

### 1. WebSocket Over HTTP Polling

**Decision**: Use persistent WebSocket connections for all game communication, not HTTP polling.

When two players make moves, both clients need to receive the updated board in milliseconds. HTTP polling—where each client repeatedly asks "anything new?"—introduces unnecessary latency and server load. With WebSockets, the server pushes updates directly.

I built a central Hub that manages all connections:

```go
type Hub struct {
    connections map[string]*Connection           // Active users
    gameRooms   map[string]map[string]*Connection // Per-game rooms
    broadcast   chan *BroadcastMessage
    mu          sync.RWMutex
}
```

When a player makes a move, the server validates it, updates state, and broadcasts to the `gameRooms[gameID]` map—instantly reaching both players. A single `sync.RWMutex` keeps concurrent access safe without the overhead of a more complex concurrency primitive.

**Tradeoff acknowledged**: This architecture doesn't horizontally scale. A second server instance wouldn't share the in-memory connection map. For horizontal scaling, the `BroadcastToGame` call would need to route through a Redis Pub/Sub channel instead of an in-process map. I documented this explicitly as a known limitation rather than pretending it away.

---

### 2. Minimax with Alpha-Beta Pruning for the AI Bot

**Decision**: Implement minimax search with alpha-beta pruning and iterative deepening, rather than using a lookup table or random play.

A bot that plays randomly isn't fun. A bot that plays perfectly at depth 9 is unbeatable and frustrating. I needed an AI that plays well within a time budget.

Minimax explores the game tree: the bot maximizes its score, the opponent minimizes it. The problem is the branching factor—Connect4 has 7 columns, so depth-7 search visits up to 7⁷ ≈ 823,000 nodes. Alpha-beta pruning eliminates branches that can't affect the final decision:

```go
if beta <= alpha {
    break // This branch won't be chosen—stop evaluating it
}
```

In practice, alpha-beta with good move ordering reduces the effective search space by ~90%, bringing depth-7 search into the hundreds of milliseconds range.

To make the bot responsive under all conditions, I added **iterative deepening**:

```go
for depth := 1; depth <= 7; depth++ {
    if time.Now().After(deadline) { break }
    bestMove = search(depth)
}
```

The bot always has a move ready to return. If the server is under load, it returns the best move found at depth 3 or 4. If the server has time, it goes deeper. This approach—progressively deeper searches with a hard timeout—was the key insight that made the bot both fast and competent.

**Move ordering** further improves pruning: center columns are evaluated first because they're statistically stronger in Connect4. This simple heuristic meaningfully improves the pruning effectiveness.

---

### 3. Session Persistence and Reconnection Handling

**Decision**: Give players a 30-second grace period to reconnect before a game is abandoned.

Network drops happen. If a player's browser closes mid-game and a disconnect immediately triggers a forfeit, the product feels brittle. But holding a game open forever while one player is gone isn't fair either.

The solution is a time-bounded grace window:

```
Player disconnects
    ↓
Server marks player as disconnected + records timestamp
    ↓
30-second countdown begins
    ├── Reconnects within 30s → Game continues normally
    └── 30s expires → Game abandoned, opponent wins
```

When a reconnecting player sends a `reconnect` WebSocket message, the handler:
1. Verifies the player belongs to that game
2. Registers the new connection in the hub's game room
3. Pushes the full current `game_state` message so the client re-renders correctly

This design means the reconnection path is stateless from the server's perspective—the game session is the source of truth, not the WebSocket connection object. A new connection can seamlessly replace an old one.

---

### 4. Kafka Analytics Pipeline—Async by Design

**Decision**: Route analytics events through Kafka asynchronously, and never block game logic waiting for them.

I wanted to track game metrics: move counts, session durations, win rates by player. A naive approach—synchronous database writes inside the move handler—would add I/O latency to every single player action.

The solution is a fire-and-forget producer in a goroutine:

```go
// In the move handler—never blocks
go func() {
    s.analyticsProducer.SendMoveMade(gameID, playerID, col, row)
}()

// Also save to local event table for immediate queries (no Kafka dependency)
s.eventRepo.Create(ctx, event)
```

This gives three properties I wanted:
- **Zero latency impact** on gameplay—Kafka is off the critical path
- **Resilience**—if Confluent Cloud is unreachable, the local `game_events` table still has the data
- **Scalability**—a separate analytics consumer processes events without coupling to the game server

The dual-write (Kafka + local table) adds some complexity, but it means the analytics pipeline can tolerate Kafka downtime without losing data.

---

### 5. Matchmaking with a Bot Fallback

**Decision**: If a player waits 10 seconds without finding a human opponent, automatically start a bot game.

Empty queues kill multiplayer games. If a player joins and waits indefinitely with no match, they leave. The bot fallback ensures every player gets a game within 10 seconds regardless of server population.

The matchmaking service runs a background worker that ticks every second:

```
Every 1 second:
    For each player in queue:
        If timeout expired AND no paired partner → create bot game
        If two players both ready → pair them, create human game
```

This makes the queue self-draining. Players with overlapping windows are paired; players without fallback to bot games. From a UX perspective, the transition is invisible—the client receives a `game_started` message either way.

---

## What I'd Do Differently

**Redis for WebSocket state**: The in-memory connection hub works for a single server, but a production system with multiple instances needs shared state. Redis Pub/Sub would replace the in-process broadcast channel.

**Optimistic UI updates**: Currently the client waits for the server's `move_made` message before rendering the disc. For snappier UX, the client could render immediately and roll back if the server rejects the move.

**Deeper bot search with transposition tables**: The minimax bot caches board positions in a transposition table, but the implementation is partial. A complete Zobrist-hash-based table would dramatically reduce redundant computation.

---

## Summary

What made this project interesting wasn't any single component—it was the intersection of real-time systems, AI, event-driven architecture, and fault tolerance in one cohesive application. Each technical decision involved a genuine tradeoff (latency vs. scalability, responsiveness vs. accuracy, simplicity vs. resilience), and working through those tradeoffs explicitly is what I find most valuable about projects like this.

The full project is at: **https://github.com/luxmikant/Connect4**  
In-depth architecture documentation: **[PROJECT_IN_DEPTH.md](PROJECT_IN_DEPTH.md)**
