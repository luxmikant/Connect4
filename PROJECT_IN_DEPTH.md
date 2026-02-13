# Connect4 Multiplayer - In-Depth Project Analysis

## 📋 Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture & Tech Stack](#architecture--tech-stack)
3. [File Structure](#file-structure)
4. [Core Components](#core-components)
5. [Game Flow](#game-flow)
6. [Design Tradeoffs](#design-tradeoffs)
7. [Key Features Implementation](#key-features-implementation)

---

## Project Overview

**Connect4 Multiplayer** is a real-time multiplayer Connect 4 game system featuring:
- Live WebSocket-based multiplayer gameplay
- AI opponent using minimax algorithm with alpha-beta pruning
- Automatic matchmaking with 10-second timeout
- Session persistence with reconnection support (30-second grace period)
- Kafka-based analytics pipeline for real-time game metrics
- PostgreSQL persistence with Redis caching
- React frontend with Vite

**Repository**: https://github.com/luxmikant/Connect4  
**Deployment**: Render (Production) + Local Docker (Development)

---

## Architecture & Tech Stack

### Backend Stack
- **Language**: Go 1.23
- **Framework**: Gin (HTTP/REST)
- **WebSocket**: Gorilla WebSocket
- **Database**: PostgreSQL (primary) + SQLite (testing)
- **ORM**: GORM
- **Message Queue**: Apache Kafka (analytics)
- **Cache**: In-memory (no external Redis in current deployment)
- **Logging**: `log/slog`

### Frontend Stack
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Animation**: Framer Motion
- **Styling**: Tailwind CSS
- **State Management**: React Hooks + Context API
- **HTTP Client**: Fetch API
- **3D Effects**: Three.js (via react-spline, react-three-fiber)
- **UI Components**: Lucide icons, Shadcn UI patterns

### Deployment
- **Backend**: Render (Go Docker)
- **Frontend**: Vercel (React SPA)
- **Database**: Supabase PostgreSQL (cloud) / Local Docker PostgreSQL
- **Kafka**: Confluent Cloud (production) / Local Docker (development)
- **Auth**: Supabase Authentication

---

## File Structure

```
project-root/
├── cmd/                           # Entry points
│   ├── server/
│   │   └── main.go               # Backend server entry
│   ├── analytics/
│   │   └── main.go               # Kafka analytics consumer
│   └── migrate/
│       └── main.go               # Database migration runner
│
├── internal/                      # Private application code
│   ├── api/
│   │   ├── handlers/             # HTTP handlers for REST endpoints
│   │   │   └── player.go         # Player auth/profile endpoints
│   │   ├── middleware/           # Auth, logging, CORS middleware
│   │   └── routes/               # Route definitions
│   │
│   ├── websocket/                # Real-time multiplayer engine
│   │   ├── hub.go                # Central WebSocket connection manager
│   │   ├── handler.go            # Message routing & business logic (968 lines)
│   │   ├── connection.go         # Individual connection management
│   │   ├── message.go            # Message type definitions (31 types)
│   │   ├── service.go            # WebSocket service initialization
│   │   ├── mocks_test.go         # Test mocks
│   │   └── integration_test.go   # Integration tests
│   │
│   ├── game/                      # Game logic engine
│   │   ├── service.go            # Game session management (1181 lines)
│   │   ├── engine.go             # Move validation & board state
│   │   ├── player_test.go        # Game logic tests
│   │   └── mocks_test.go         # Test mocks
│   │
│   ├── bot/                       # AI opponent engine
│   │   ├── minimax.go            # Alpha-beta pruning algorithm (518 lines)
│   │   │   ├── GetBestMove()
│   │   │   ├── GetBestMoveWithTimeout()
│   │   │   ├── minimax()         # Core minimax implementation
│   │   │   ├── minimaxWithDeadline()
│   │   │   ├── EvaluatePosition()
│   │   │   ├── evaluateWindows() # 4-in-a-row pattern detection
│   │   │   └── evaluateCenterControl()
│   │   ├── service.go            # Bot game orchestration
│   │   ├── player.go             # Bot player logic
│   │   └── minimax_test.go       # AI tests
│   │
│   ├── matchmaking/              # Queue-based matchmaking
│   │   ├── service.go            # Matchmaking engine (434 lines)
│   │   │   ├── JoinQueue()
│   │   │   ├── LeaveQueue()
│   │   │   ├── StartMatchmaking()
│   │   │   └── Match timeout: 10 seconds → bot fallback
│   │   ├── mocks_test.go         # Test mocks
│   │   └── service_property_test.go
│   │
│   ├── analytics/                 # Kafka event producer
│   │   ├── service.go            # Analytics orchestration
│   │   ├── producer.go           # Kafka producer
│   │   └── producer_property_test.go
│   │
│   ├── auth/                      # Supabase authentication
│   │   └── supabase.go           # JWT validation & user profile
│   │
│   ├── config/                    # Configuration management
│   │   └── config.go             # Viper-based config loading
│   │
│   ├── database/                  # Database layer
│   │   ├── database.go           # Connection & initialization
│   │   ├── migrator.go           # Migration runner
│   │   └── repositories/         # Data access objects
│   │       ├── game_session_repository.go    # GameSession CRUD
│   │       ├── player_repository.go          # Player CRUD
│   │       ├── move_repository.go            # Move history CRUD
│   │       ├── player_stats_repository.go    # Leaderboard queries
│   │       └── game_event_repository.go      # Event logging
│   │
│   └── stats/                     # Player statistics
│       └── service.go            # Win rate, ranking calculations
│
├── pkg/                           # Public libraries
│   └── models/                    # Core domain models
│       ├── game.go               # GameSession, Board, GameStatus
│       ├── player.go             # Player, PlayerColor (Red/Yellow)
│       ├── stats.go              # PlayerStats, leaderboard data
│       ├── move.go               # Move history record
│       ├── events.go             # GameEvent types
│       ├── errors.go             # Custom error definitions
│       └── utils.go              # Helper functions
│
├── migrations/                    # SQL migration files
│   ├── 001_create_players_table.sql
│   ├── 002_create_game_sessions_table.sql
│   ├── 003_create_moves_table.sql
│   ├── 004_create_player_stats_table.sql
│   ├── 005_create_game_events_table.sql
│   ├── 006_add_game_session_indexes.sql
│   ├── 007_create_analytics_snapshots_table.sql
│   ├── 008_create_profiles.sql
│   ├── 009_link_players_to_auth.sql
│   ├── 010_create_upsert_player_function.sql
│   └── 011_add_custom_room_columns.sql
│
├── web/                           # React frontend
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Landing.tsx       # MetaMask-style landing
│   │   │   ├── Lobby.tsx         # Game mode selection
│   │   │   ├── Game.tsx          # Main gameplay component
│   │   │   └── Leaderboard.tsx   # Rankings display
│   │   │
│   │   ├── components/
│   │   │   ├── GameBoard.tsx     # 6x7 grid component
│   │   │   ├── PlayerPanel.tsx   # Player info display
│   │   │   └── ...
│   │   │
│   │   ├── hooks/
│   │   │   ├── useGame.ts        # Game state management
│   │   │   ├── usePlayer.ts      # Player context
│   │   │   ├── useWebSocket.ts   # WebSocket connection
│   │   │   └── useGameSound.ts   # Sound effects
│   │   │
│   │   ├── services/
│   │   │   ├── websocket.ts      # WebSocket client
│   │   │   ├── playerService.ts  # Player API calls
│   │   │   └── apiService.ts     # General API client
│   │   │
│   │   ├── types/
│   │   │   └── websocket.ts      # Message type definitions
│   │   │
│   │   ├── contexts/
│   │   │   ├── AuthContext.tsx   # Supabase auth
│   │   │   └── PlayerContext.tsx # Player state
│   │   │
│   │   └── App.tsx               # Router & layout
│   │
│   ├── public/
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   └── package.json
│
├── docker-compose.yml            # Local development stack
├── Dockerfile.server             # Backend container
├── Dockerfile.analytics          # Analytics consumer container
├── Makefile                      # Development commands
├── go.mod / go.sum              # Go dependencies
└── migrations/                   # SQL migration files
```

---

## Core Components

### 1. WebSocket Hub (Real-Time Engine)

**File**: `internal/websocket/hub.go`

```go
type Hub struct {
    connections map[string]*Connection      // Active user connections
    gameRooms   map[string]map[string]*Connection  // Game-specific rooms
    broadcast   chan *BroadcastMessage     // Message broadcast queue
    mu          sync.RWMutex               // Thread-safe access
}
```

**Key Operations**:
- `RegisterConnection()` - Add new WebSocket client
- `BroadcastToGame(gameID)` - Send message to all players in a game
- `UpdateConnectionUserID()` - Handle player authentication
- `addToGameRoom()` - Track players in active game rooms
- `removeFromGameRoom()` - Clean up when players disconnect

**Thread Safety**: RWMutex protects concurrent access to connections map

---

### 2. Game Service (Business Logic)

**File**: `internal/game/service.go` (1181 lines)

**Responsibilities**:
- Session lifecycle: Create, retrieve, end games
- Custom room management: Create room → room code → join room → rematch
- Player color assignment (Red/Yellow)
- Disconnection handling with 30-second grace period
- Move validation and board state management
- Win/draw detection
- Analytics event publishing

**Key Methods**:
```go
CreateSession(player1, player2 string) (*GameSession, error)
CreateCustomRoom(creator string) (*GameSession, roomCode string, error)
JoinCustomRoom(roomCode, username string) (*GameSession, error)
RematchCustomRoom(gameID, username string) (*GameSession, error)
MarkPlayerDisconnected(gameID, username string) error
MarkPlayerReconnected(gameID, username string) error
CompleteGame(gameID string, winner *PlayerColor) error
```

**In-Memory Cache**:
```go
sessionCache map[string]*cachedSession  // Fast lookups
disconnectedPlayers map[string]map[string]time.Time  // Track who disconnected
```

---

### 3. AI Bot Engine (Minimax + Alpha-Beta Pruning)

**File**: `internal/bot/minimax.go` (518 lines)

**Algorithm**:
```
GetBestMove(board, player, depth)
├─ Check immediate win
├─ Check opponent's winning move (block)
└─ Use minimax with alpha-beta pruning

minimax(board, depth, alpha, beta, isMaximizing)
├─ Terminal state detection (win/lose/draw)
├─ Depth limit check
├─ Maximize: best score for bot
│  └─ For each valid move:
│     ├─ Make move on board copy
│     ├─ Recursively minimize opponent
│     ├─ Update alpha = max(alpha, score)
│     └─ Beta cutoff: if beta <= alpha, PRUNE
└─ Minimize: best score for opponent (same pattern, inverted)
```

**Optimization Techniques**:
1. **Move Ordering** (center first): `[3, 2, 4, 1, 5, 0, 6]`
   - Center columns are more valuable
   - Improves pruning effectiveness

2. **Alpha-Beta Pruning**: Eliminates ~90% of nodes
   ```go
   if beta <= alpha {
       break  // Alpha/Beta cutoff - no need to explore further
   }
   ```

3. **Evaluation Function**:
   - 4-in-a-row detection: +100,000 (win) / -100,000 (loss)
   - 3-in-a-row (open): +100 / -100
   - 2-in-a-row (open): +10 / -10
   - Center control bonus: +3 per piece

4. **Iterative Deepening with Timeout**:
   ```go
   for depth := 1; depth <= 7; depth++ {
       if time.Now().After(deadline) { break }
       move := search(depth)
   }
   ```
   - Progressively deeper searches
   - Respects time constraints
   - Always returns best-found move

5. **Transposition Table** (caching):
   ```go
   transpositionTable map[string]int  // Board position -> score
   ```

---

### 4. Matchmaking Service

**File**: `internal/matchmaking/service.go` (434 lines)

**Queue System**:
```
Player joins queue
    ↓
Wait for match (max 10 seconds)
    ├─ Found another player? → Create game
    └─ Timeout? → Create bot game instead
```

**Implementation**:
```go
type matchmakingService struct {
    queue []*QueueEntry        // Waiting players
    queueMutex sync.RWMutex    // Thread-safe access
    playerIndex map[string]int // Quick lookup
    matchTimeout 10 * time.Second
    matchInterval 1 * time.Second
}
```

**Matching Algorithm**:
1. Collect all queue entries whose timeout expires first
2. Pair them (A with B, C with D, etc.)
3. For unpaired players: create bot games after timeout

---

### 5. Database Layer (GORM + Repositories)

**Files**: `internal/database/repositories/`

**Schema**:
```sql
players
├─ id, username, auth_user_id, is_guest
├─ created_at, updated_at

game_sessions
├─ id, player1, player2, board (JSONB), current_turn
├─ status (waiting/in_progress/completed/abandoned)
├─ winner, start_time, end_time
├─ room_code (custom rooms), is_custom, created_by
├─ indexes: status, room_code, is_custom

moves
├─ id, game_id, player_id, column, row, move_number, timestamp

player_stats
├─ player_id, games_played, games_won, win_rate
├─ avg_game_duration, last_game_time

game_events
├─ id, event_type, game_id, player_id, timestamp, metadata (JSONB)

profiles (Supabase auth integration)
├─ id, username, avatar_url, updated_at
```

**Repository Pattern**:
```go
type GameSessionRepository interface {
    Create(ctx, session) error
    GetByID(ctx, id) (*GameSession, error)
    Update(ctx, session) error
    Delete(ctx, id) error
    GetByRoomCode(ctx, code) (*GameSession, error)
    FindActive(ctx) ([]*GameSession, error)
}
```

---

### 6. Analytics Pipeline (Kafka)

**Files**: `internal/analytics/producer.go`

**Events Tracked**:
- `game_started`: Player1, Player2, timestamp
- `move_made`: Column, row, move number
- `game_completed`: Winner, loser, duration
- `player_joined`, `player_disconnected`, `player_reconnected`

**Flow**:
```
Game Event
    ↓
analyticsProducer.Send*()
    ↓
Kafka Topic (e.g., "game-events")
    ↓
Analytics Consumer (cmd/analytics)
    ↓
Compute metrics & store
```

---

### 7. Frontend Architecture (React + TypeScript)

**State Management Hierarchy**:
```
App.tsx
├─ AuthContext (Supabase auth)
│   └─ user, profile, isAuthenticated
├─ PlayerContext (usePlayer hook)
│   └─ username, setUsername
└─ Page Routes
    ├─ Landing → MetaMask-style onboarding
    ├─ Lobby → Game mode selection
    │   ├─ Matchmaking queue
    │   ├─ Bot game
    │   ├─ Custom room (create/join)
    └─ Game → Gameplay
        ├─ useGame hook (game state)
        ├─ useWebSocket hook (real-time)
        └─ GameBoard component (UI)
```

**WebSocket Message Types** (31 total):

*Client → Server*:
- `join_queue`, `leave_queue` - Matchmaking
- `play_with_bot` - Bot game
- `create_custom_room`, `join_custom_room`, `rematch_custom_room` - Custom rooms
- `make_move` - Gameplay
- `reconnect` - Reconnection after disconnect
- `ping` - Keep-alive

*Server → Client*:
- `queue_joined`, `queue_status`, `match_found` - Matchmaking responses
- `room_created`, `waiting_for_opponent` - Custom room responses
- `game_started` - Game begins
- `move_made` - Opponent's move
- `game_ended` - Game completion
- `game_state` - Board state sync
- `player_joined`, `player_left` - Presence updates
- `error` - Error messages

---

## Game Flow

### 1. Matchmaking Game Flow

```
┌─── LOBBY PAGE ───────────────────────────────────┐
│ User enters username & selects "Matchmaking"     │
│ → Click "Find Opponent"                          │
└─────────────────────────────────────────────────┘
         ↓
┌─── WEBSOCKET CONNECT ───────────────────────────┐
│ wsService.connect(username)                      │
│ → GET /ws?userId=username                        │
│ → Connection registered in hub                   │
└─────────────────────────────────────────────────┘
         ↓
┌─── SEND JOIN_QUEUE MESSAGE ─────────────────────┐
│ wsService.send(MessageType.JoinQueue, {username})│
│ → handleJoinQueue() in handler.go                │
└─────────────────────────────────────────────────┘
         ↓
┌─── MATCHMAKING SERVICE ─────────────────────────┐
│ matchmakingService.JoinQueue(username)           │
│ → Add to queue with 10-second timeout            │
│ → Send queue_joined message to client            │
│                                                  │
│ [Matchmaking worker runs every 1 second]         │
│ ├─ Check if timeout expired                      │
│ ├─ Find matches (pair players)                   │
│ ├─ No match? Create bot game after timeout       │
│ └─ Match found? Create game session              │
└─────────────────────────────────────────────────┘
         ↓
      [TWO PATHS]
         ├─────────────────────────┬──────────────────────┐
         ↓ (Match found)            ↓ (Timeout, no match) │
                                                          │
    ┌──────────────────────┐    ┌─────────────────────┐  │
    │ GAME CREATED         │    │ BOT GAME CREATED    │  │
    │ Player1: userA       │    │ Player1: user       │  │
    │ Player2: userB       │    │ Player2: bot        │  │
    │ Callback triggers    │    │ Callback triggers   │  │
    └──────────────────────┘    └─────────────────────┘  │
         ↓                                   ↓            │
    ┌──────────────────────────────────────────────────┐ │
    │ Handler: onGameCreated() / onBotGameCreated()     │ │
    │ ├─ Update connection.gameID                       │ │
    │ ├─ Add connection to hub.gameRooms[gameID]        │ │
    │ ├─ Send game_started message to both players      │ │
    │ └─ Notify players of opponent name & color       │ │
    └──────────────────────────────────────────────────┘ │
         ↓                                                │
    ┌──────────────────────────────────────────────────┐ │
    │ GAME PAGE LOADED                                 │ │
    │ ├─ Render 6x7 board                              │ │
    │ ├─ Show opponent name & your color               │ │
    │ └─ Wait for first player's turn                  │ │
    └──────────────────────────────────────────────────┘ │
         ↓
    [GAMEPLAY LOOP - See section 2]
```

### 2. Gameplay Flow

```
┌─── GAME IN PROGRESS ────────────────────────────┐
│ Player turn indicator: "Your turn" / "Waiting..." │
└─────────────────────────────────────────────────┘
         ↓
┌─── USER CLICKS COLUMN ──────────────────────────┐
│ (Only enabled if isMyTurn === true)             │
│ → Call makeMove(column)                          │
└─────────────────────────────────────────────────┘
         ↓
┌─── SEND MAKE_MOVE MESSAGE ──────────────────────┐
│ wsService.send(MessageType.MakeMove, {           │
│     gameId: gameState.id,                        │
│     column: selectedColumn                       │
│ })                                               │
└─────────────────────────────────────────────────┘
         ↓
┌─── BACKEND VALIDATION ──────────────────────────┐
│ handleMakeMove() in handler.go                   │
│ ├─ Verify player's turn                          │
│ ├─ Validate move (column valid & not full)       │
│ ├─ Get current game session                      │
│ ├─ engine.MakeMove(column, playerColor)          │
│ │   └─ Update board.Grid[row][column]            │
│ │   └─ Update board.Height[column]++             │
│ ├─ Check for win: board.CheckWin()               │
│ │   ├─ Horizontal 4-in-a-row?                    │
│ │   ├─ Vertical 4-in-a-row?                      │
│ │   ├─ Diagonal 4-in-a-row?                      │
│ │   └─ Return winner or nil                      │
│ ├─ Save move to database: moveRepo.Create()      │
│ ├─ Switch turn: service.SwitchTurn()             │
│ └─ Send move_made message to both players        │
└─────────────────────────────────────────────────┘
         ↓
   [THREE PATHS]
   ├─────────────────┬──────────────────┬──────────────┐
   │ (Human player)  │ (No winner/draw) │ (Someone won)│
   │                 │                  │              │
   ↓ BOT MOVE        ↓ NEXT TURN         ↓ GAME ENDED  │
   │                                                   │
   ├─ Get bot best  ├─ Turn switches    ├─ Determine  │
   │  move using    │  to opponent      │  winner     │
   │  minimax()     │                   │  (player1   │
   │                ├─ Send turn_change │   or        │
   │ ├─ Immediate   │  message         │   player2)  │
   │ │  win? Play   │                   │             │
   │ │  that move   ├─ Frontend updates ├─ Update     │
   │ │              │  button states    │  player_    │
   │ ├─ Need to     │  (your turn or    │  stats      │
   │ │  block? Play │   waiting)        │  (wins,     │
   │ │  that move   │                   │   losses)   │
   │ │              │                   │             │
   │ └─ Alpha-beta  │                   ├─ Send       │
   │    search      │                   │  game_ended │
   │    depth 5     │                   │  message    │
   │                │                   │             │
   ├─ Make move    ├─ Opponent gets    ├─ Show       │
   │  on board     │  their turn       │  "You won!" │
   │              │  (if human)        │  or         │
   ├─ Switch to   │                   │  "You lost!"│
   │  opponent     └─ Loop until game  │             │
   │                 ends              ├─ Offer      │
   └─ Send                             │  "Play      │
     move_made                         │  Again"     │
     message                           │  button     │
                                       │             │
                                       └─ Analytics  │
                                         sent to     │
                                         Kafka       │
```

### 3. Custom Room Flow

```
┌─── LOBBY PAGE ──────────────────────┐
│ User selects "Custom Room"           │
│ ├─ CREATE: Click "Create Room"       │
│ └─ JOIN: Enter room code + click     │
└─────────────────────────────────────┘
         ↓
    [TWO PATHS]
    ├──────────────────────┬──────────────────┐
    │ CREATE ROOM          │ JOIN ROOM         │
    └─────────────────────┘──────────────────┘
         ↓
    ┌────────────────────────────────────────┐
    │ WebSocket connect                      │
    └────────────────────────────────────────┘
         ↓
    [PATH A: CREATE]          [PATH B: JOIN]
    │                         │
    ├─ wsService.send(        ├─ wsService.send(
    │  CreateCustomRoom,      │  JoinCustomRoom,
    │  {username}             │  {username, roomCode}
    │ )                       │ )
    │                         │
    └─ handleCreateCustom     └─ handleJoinCustom
      Room()                    Room()
         ↓                         ↓
    ├─ Generate 8-char   ├─ Look up room by code
    │  room code         │
    │  (e.g., VWKYXL8L)  ├─ Verify room exists
    │                    │  & is waiting
    ├─ Create game       │
    │  session:          ├─ Add joiner as Player2
    │  - Player1: user   │
    │  - Player2: "wait" ├─ Change status:
    │  - status: waiting │  waiting → in_progress
    │  - room_code: code │
    │  - is_custom: true ├─ Initialize board
    │  - created_by: user│
    │                    ├─ Set start time
    ├─ Add to game room  │
    │  hub.gameRooms[]   ├─ Save move & send
    │                    │  game_started to both
    ├─ Send room_created │
    │  message with code ├─ Handle reconnection:
    │  & URL             │  If creator reconnects,
    │                    │  JoinCustomRoom returns
    ├─ Show room code    │  session, sends
    │  & "Waiting for    │  waiting_for_opponent
    │  opponent..." UI   │  message
    │
    └─ Both players wait
       until opponent
       joins
         │
         ├─ Creator can share code
         │ (copy/paste or QR code)
         │
         └─ Opponent scans QR or
            enters code → joins
               ↓
         [GAME STARTS]
         ├─ Notify both players
         ├─ Initialize board
         └─ Begin gameplay loop
            (see section 2)
               ↓
         [GAME ENDS]
         ├─ Determine winner
         ├─ Send game_ended
         └─ Show "Play Again"
            button
               ↓
         [REMATCH FLOW]
         ├─ wsService.send(
         │  RematchCustomRoom,
         │  {gameId, username}
         │ )
         │
         ├─ handleRematchCustomRoom()
         │  ├─ Get old session
         │  ├─ Clear room_code from
         │  │  old session (preserve history)
         │  ├─ Create NEW session:
         │  │  - Same players
         │  │  - Same room code
         │  │  - Fresh board
         │  ├─ Move players to new
         │  │  game room in hub
         │  └─ Send game_started
         │
         └─ NEW GAME STARTS
            (loop back to gameplay)
```

### 4. Reconnection Flow

```
┌──────────────────────────────┐
│ Player in active game        │
│ Internet connection lost      │
│ WebSocket closes             │
└──────────────────────────────┘
         ↓
┌──────────────────────────────┐
│ Frontend detects disconnect  │
│ (ws.onclose event)           │
│ Auto-reconnect in 3 seconds  │
└──────────────────────────────┘
         ↓
┌──────────────────────────────┐
│ NEW WebSocket connection     │
│ Server creates new           │
│ Connection object            │
│ (old one discarded)          │
└──────────────────────────────┘
         ↓
┌──────────────────────────────┐
│ Send reconnect message       │
│ wsService.send(              │
│   MessageType.Reconnect,     │
│   {gameId, username}         │
│ )                            │
└──────────────────────────────┘
         ↓
┌──────────────────────────────┐
│ handleReconnect()            │
│ ├─ Verify game exists        │
│ ├─ Verify user in game       │
│ ├─ Update connection.gameID  │
│ ├─ Add to hub.gameRooms[]    │
│ └─ Send current game_state   │
│   message to client          │
└──────────────────────────────┘
         ↓
┌──────────────────────────────┐
│ [TIMEOUT: 30 seconds]        │
│ ├─ If still disconnected     │
│ │ └─ Abandon game after 30s  │
│ │   (grace period expires)   │
│ └─ If reconnected before 30s │
│   └─ Continue game normally  │
└──────────────────────────────┘
         ↓
[GAME CONTINUES OR ABANDONED]
```

---

## Design Tradeoffs

### 1. **In-Memory Cache vs Redis**

**Tradeoff Made**: In-memory cache (simple map + mutex)

**Pros**:
- ✅ Zero network latency
- ✅ Simpler deployment (no Redis service)
- ✅ Sufficient for single-server deployment

**Cons**:
- ❌ Not shared across multiple server instances
- ❌ Data lost on server restart
- ❌ Can cause memory bloat if not cleaned up

**Recommendation**: Use Redis for production multi-instance deployments

---

### 2. **Synchronous vs Asynchronous Game State**

**Tradeoff Made**: Synchronous message flow

```go
// Synchronous: Wait for response
client sends move
    ↓ (waits)
server processes
server broadcasts to all players
all clients receive at same time
```

**Pros**:
- ✅ Consistent game state across all players
- ✅ No race conditions
- ✅ Simpler debugging

**Cons**:
- ❌ Slower response time
- ❌ Blocked if server slow

**Alternative (not used)**: Async with eventual consistency
- Client optimistically updates board
- Server validates & broadcasts
- Rollback if invalid

---

### 3. **Minimax Depth vs Move Quality**

**Tradeoff Made**: Depth 5 with timeout constraints

```
Depth 1: ~7 nodes (immediate)
Depth 3: ~7³ ≈ 343 nodes (very fast)
Depth 5: ~7⁵ ≈ 16K nodes (with pruning: ~1.5K, ~100ms)
Depth 7: ~7⁷ ≈ 823K nodes (with pruning: ~50K, ~1-2s)
Depth 9: ~7⁹ ≈ 40M nodes (too slow)
```

**Implemented**: 
```go
for depth := 1; depth <= 7; depth++ {
    if time.Now().After(deadline) { break }
    bestMove = search(depth)
}
```

**Pros**:
- ✅ Responsive gameplay (~500ms timeout)
- ✅ Good move quality
- ✅ Handles slow networks

**Cons**:
- ❌ Not perfect play (depth 9 would solve game)
- ❌ Occasionally missable wins at depth 5

**Improvement**: Use transposition table (partially implemented)

---

### 4. **WebSocket Hub Broadcast vs Direct Routing**

**Tradeoff Made**: Hub-based broadcast with connection tracking

```go
// Current approach
handler.hub.BroadcastToGame(gameID, message, excludeUserID)
    ↓
hub looks up gameRooms[gameID]
    ↓
sends to each connection

// Alternative: Direct routing (not used)
handler.conn.Send(message)  // Only to one player
```

**Pros**:
- ✅ Scales to many connections
- ✅ Centralized connection management
- ✅ Easy to implement features like spectating

**Cons**:
- ❌ Single point of failure
- ❌ Can't horizontally scale (no pub/sub)

**For horizontal scaling**: Replace with Redis Pub/Sub or message queue

---

### 5. **Move Validation on Client vs Server**

**Tradeoff Made**: Validation on both (client + server)

**Client-side** (for instant feedback):
```typescript
if (!isMyTurn) return;  // Don't allow clicks
if (!board.isValidMove(col)) return;  // Grayed out
```

**Server-side** (for security):
```go
if session.GetCurrentPlayer() != username { return error }
if !board.IsValidMove(col) { return error }
```

**Pros**:
- ✅ Prevents cheating (server is source of truth)
- ✅ Good UX (instant feedback on client)
- ✅ Fails safely if network lag

**Cons**:
- ❌ Code duplication
- ❌ Potential inconsistencies

---

### 6. **Kafka Analytics (Async) vs Sync Logging**

**Tradeoff Made**: Async Kafka events + local event table

```go
// Async - don't block game flow
go func() {
    s.analyticsProducer.SendGameStarted(...)
}()

// Also save locally for immediate queries
s.eventRepo.Create(ctx, event)
```

**Pros**:
- ✅ Never blocks game logic
- ✅ Survives Kafka downtime (local events)
- ✅ Batch processing possible
- ✅ Real-time dashboards

**Cons**:
- ❌ At-least-once delivery (some duplication)
- ❌ Complexity of managing two stores

---

### 7. **Database Normalization vs Denormalization**

**Tradeoff Made**: Highly normalized + JSON columns for denormalization

```sql
-- Normalized: game_sessions.board (JSON)
-- Avoids 7*6 = 42 board_cell records per game

-- Normalized: game_sessions references players
-- But stores player1, player2 as strings

-- Denormalized: game_sessions includes winner
-- Could join to moves table, but stored directly
```

**Pros**:
- ✅ Balanced approach
- ✅ Easy queries
- ✅ JSON fast to serialize/deserialize

**Cons**:
- ❌ Query flexibility reduced
- ❌ No relational constraints on board

---

## Key Features Implementation

### Feature 1: Real-Time Multiplayer

**Components**:
1. WebSocket connection (Gorilla WebSocket)
2. Hub for connection management
3. Message queue for broadcasts
4. Atomic game state updates

**Code Flow**:
```
Client → WebSocket → Handler → GameService → Database
                        ↓
                      Hub.Broadcast
                        ↓
                    All connected clients
```

---

### Feature 2: Custom Rooms

**Components**:
1. Room code generation (8 alphanumeric)
2. Database `room_code` unique index
3. `JoinCustomRoom()` allows creator to rejoin

**Code Flow**:
```
1. Creator clicks "Create"
   → gameService.CreateCustomRoom("user1")
   → Generates room code "VWKYXL8L"
   → Creates session with Player2="waiting"
   → Returns room code to frontend

2. Opponent enters code & clicks "Join"
   → gameService.JoinCustomRoom("VWKYXL8L", "user2")
   → Updates session.Player2="user2"
   → Changes status: waiting → in_progress
   → Returns session

3. Creator disconnects & reconnects
   → gameService.JoinCustomRoom("VWKYXL8L", "user1")
   → Detects creator trying to rejoin
   → Returns session (allows reconnection)
   → Frontend receives waiting_for_opponent message
```

---

### Feature 3: Rematch System

**Components**:
1. `RematchCustomRoom()` creates new session
2. Old session.RoomCode set to NULL (preserve history)
3. New session gets same room code
4. Hub moves players to new game room

**Code Flow**:
```
Game 1 ends → Player1 clicks "Play Again"
    ↓
handleRematchCustomRoom(gameID="game-1", username="user1")
    ↓
1. Get game-1 session
   - Player1="user1", Player2="user2"
   - RoomCode="VWKYXL8L"

2. Update game-1
   - RoomCode = NULL
   - Status = completed (already was)

3. Create game-2
   - Player1="user1", Player2="user2"
   - RoomCode="VWKYXL8L"
   - Status="in_progress"
   - Fresh board

4. Move hub connections
   - Remove from hub.gameRooms["game-1"]
   - Add to hub.gameRooms["game-2"]

5. Send game_started to both players
    ↓
New game begins with same players & room code
```

---

### Feature 4: Alpha-Beta Pruning

**Implementation Details**:

```go
minimax(board, depth, alpha, beta, isMaximizing):
    
    // Base cases
    if depth == 0: return evaluate(board)
    if winner: return scoreWin or scoreLose
    
    if isMaximizing:
        maxScore = -∞
        for each valid move:
            score = minimax(..., depth-1, alpha, beta, false)
            maxScore = max(maxScore, score)
            alpha = max(alpha, score)
            
            if beta <= alpha:
                break  // PRUNE: opponent won't let us reach here
        
        return maxScore
    
    else:
        minScore = +∞
        for each valid move:
            score = minimax(..., depth-1, alpha, beta, true)
            minScore = min(minScore, score)
            beta = min(beta, score)
            
            if beta <= alpha:
                break  // PRUNE: we already found better move
        
        return minScore
```

**Pruning Example**:
```
max node alpha=3, beta=∞
  └─ try move A: returns score 5
    └─ alpha=5 (5 > 3)
  
  └─ try move B: (need to evaluate)
    └─ min node with alpha=5, beta=∞
      └─ try opponent move 1: returns 4
        └─ beta=4 (4 < ∞)
      
      └─ PRUNE: beta(4) <= alpha(5)?
        └─ YES! Don't evaluate more opponent moves
        └─ Return 4
    
  └─ move B score is 4 < 5
  
  └─ try move C: (potentially pruned earlier)
    └─ ...similar pattern...

Result: Evaluated ~3-4 moves instead of 7 per node
```

---

### Feature 5: Analytics Pipeline

**Kafka Topics**:
1. `game-events` - Raw game events
2. `game-metrics` - Aggregated statistics

**Consumer** (`cmd/analytics/main.go`):
```go
Events consumed:
- game_started: Track unique players, game count
- move_made: Move analysis, average moves per game
- game_completed: Winner tracking, game duration
- player_joined/left/reconnected: Presence analysis

Computed metrics:
- Games per hour
- Average game duration
- Player win rates
- Peak concurrent players
- ...more
```

**Not implemented yet**:
- Snapshots table storage
- Real-time dashboards
- Time-series queries

---

### Feature 6: Session Persistence & Reconnection

**Disconnection Tracking**:
```go
type gameService struct {
    disconnectedPlayers map[string]map[string]time.Time
    //                  gameID → username → disconnect time
    disconnectTimeout time.Duration // 30 seconds
}
```

**Reconnection Process**:
```
1. Player connection closes
   → Mark in disconnectedPlayers[gameID][username]
   → Record disconnect timestamp

2. Player reconnects within 30 seconds
   → handleReconnect() called
   → Call MarkPlayerReconnected()
   → Remove from disconnectedPlayers
   → Send current game_state
   → Game continues

3. 30 seconds pass without reconnect
   → Cleanup worker finds expired entry
   → Call HandleDisconnectionTimeout()
   → End game or continue with remaining player
   → Clean up entry
```

---

## Summary Table

| Feature | Location | Tech | Complexity |
|---------|----------|------|------------|
| Real-time multiplayer | `websocket/` | Gorilla WS | High |
| Game logic | `game/engine.go` | Pure Go | Medium |
| AI bot | `bot/minimax.go` | Alpha-beta | High |
| Matchmaking | `matchmaking/service.go` | Queue + timeout | Medium |
| Database | `database/repositories/` | GORM + PostgreSQL | Medium |
| Analytics | `analytics/producer.go` | Kafka | High |
| Frontend | `web/src/` | React + TypeScript | High |
| Custom rooms | `game/service.go` | Room codes + DB | Medium |
| Reconnection | `game/service.go` | Timestamps | Low |

---

**END OF ANALYSIS**

This document captures the entire architecture, component details, game flows, design tradeoffs, and implementation specifics of the Connect4 multiplayer system. Use this as a reference for:
- Onboarding new developers
- Architecture discussions
- Feature additions
- Performance optimization
- Deployment decisions
