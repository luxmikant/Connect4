# Test Documentation

This document provides an overview of all test files in the Connect 4 Multiplayer project, including their structure, purpose, and the functionality they validate.

## Test Overview

| Test File | Type | Package | Description |
|-----------|------|---------|-------------|
| `internal/game/service_test.go` | Unit | game | Game service session management tests |
| `internal/game/engine_property_test.go` | Property | game_test | Game engine property-based tests |
| `internal/database/repositories/player_repository_test.go` | Unit | repositories_test | Player repository CRUD tests |
| `internal/database/repositories/game_session_repository_test.go` | Unit | repositories_test | Game session repository tests |
| `internal/database/repositories/player_stats_repository_test.go` | Unit | repositories_test | Player statistics repository tests |
| `internal/database/repositories/manager_test.go` | Unit | repositories_test | Repository manager integration tests |
| `pkg/models/basic_test.go` | Unit | models_test | Basic model functionality tests |
| `pkg/models/persistence_property_test.go` | Property | models_test | Model persistence property tests |

---

## Game Service Tests

**File:** `internal/game/service_test.go`

**Purpose:** Tests the GameService which manages game session lifecycle, turn management, player disconnection handling, and caching.

### Test Structure

```
TestCreateSession
├── creates session with valid players
├── fails with empty player1
└── fails with same players

TestAssignPlayerColors
└── assigns red to player1 and yellow to player2

TestGetCurrentTurn
├── returns player1 when red's turn
└── returns player2 when yellow's turn

TestSwitchTurn
├── switches from red to yellow
└── fails for inactive game

TestCompleteGame
├── completes game with winner
└── fails for inactive game

TestPlayerDisconnection
├── marks player as disconnected
├── marks player as reconnected
└── fails for player not in game

TestGetDisconnectionTimeRemaining
├── returns remaining time for disconnected player
└── returns zero for connected player

TestCacheOperations
├── caches and retrieves session
├── invalidates cache
└── returns cache stats
```

### Mock Implementations
- `MockGameSessionRepository` - Mocks game session database operations
- `MockPlayerStatsRepository` - Mocks player statistics operations
- `MockMoveRepository` - Mocks move storage operations
- `MockGameEventRepository` - Mocks analytics event operations

---

## Game Engine Property Tests

**File:** `internal/game/engine_property_test.go`

**Build Tag:** `property`

**Purpose:** Property-based tests validating game logic invariants across randomly generated inputs.

### Test Structure

```
TestMoveValidationProperty (Property 7)
├── valid moves should be accepted for non-full columns
├── negative columns should be rejected
├── columns >= 7 should be rejected
├── moves should place discs in lowest available position
└── players can only move on their turn

TestWinAndDrawDetectionProperty (Property 8)
├── horizontal 4-in-a-row should be detected as win
├── vertical 4-in-a-row should be detected as win
├── diagonal 4-in-a-row (TL-BR) should be detected as win
├── diagonal 4-in-a-row (TR-BL) should be detected as win
├── full board without winner should be detected as draw
├── non-full boards without 4-in-a-row should not detect winner or draw
├── game engine correctly identifies win conditions
└── game engine correctly identifies draw conditions
```

### Requirements Validated
- **Property 7:** Game Move Validation and Physics (Requirements 5.1, 5.2)
- **Property 8:** Win and Draw Detection (Requirements 5.3, 5.4)

---

## Player Repository Tests

**File:** `internal/database/repositories/player_repository_test.go`

**Purpose:** Tests CRUD operations for the Player entity using an in-memory SQLite database.

### Test Structure

```
PlayerRepositoryTestSuite
├── TestCreate_Success
├── TestCreate_NilPlayer
├── TestGetByID_Success
├── TestGetByID_NotFound
├── TestGetByID_EmptyID
├── TestGetByUsername_Success
├── TestUpdate_Success
├── TestDelete_Success
├── TestList_Success
└── TestList_WithPagination
```

### Setup/Teardown
- Creates in-memory SQLite database per test
- Auto-migrates Player schema
- Closes database connection after each test

---

## Game Session Repository Tests

**File:** `internal/database/repositories/game_session_repository_test.go`

**Purpose:** Tests game session persistence including active game queries and player lookups.

### Test Structure

```
GameSessionRepositoryTestSuite
├── TestCreate_Success
├── TestCreate_NilGameSession
├── TestGetByID_Success
├── TestGetByID_NotFound
├── TestUpdate_Success
├── TestGetActiveGames_Success
├── TestGetGamesByPlayer_Success
└── TestGetGameHistory_Success
```

### Key Validations
- Session creation with board initialization
- Status filtering (waiting, in_progress, completed)
- Player-based game lookups
- Game history pagination

---

## Player Stats Repository Tests

**File:** `internal/database/repositories/player_stats_repository_test.go`

**Purpose:** Tests player statistics tracking, leaderboard queries, and game stats updates.

### Test Structure

```
PlayerStatsRepositoryTestSuite
├── TestCreate_Success
├── TestGetByUsername_Success
├── TestGetByUsername_NotFound
├── TestUpdate_Success
├── TestGetLeaderboard_Success
├── TestUpdateGameStats_NewPlayer
├── TestUpdateGameStats_ExistingPlayer
└── TestUpdateGameStats_EmptyUsername
```

### Key Validations
- Win rate calculation accuracy
- Average game time updates
- Leaderboard ordering (by wins descending)
- Atomic stats updates for new/existing players

---

## Repository Manager Tests

**File:** `internal/database/repositories/manager_test.go`

**Purpose:** Tests the repository manager which coordinates all repositories and handles transactions.

### Test Structure

```
ManagerTestSuite
├── TestNewManager_Success
├── TestHealthCheck_Success
├── TestGetConnectionStats_Success
├── TestWithTransaction_Success
├── TestWithTransaction_Rollback
├── TestBeginTransaction_Success
└── TestRepositoryIntegration
```

### Key Validations
- All repositories properly initialized
- Database health checks
- Connection pool statistics
- Transaction commit/rollback behavior
- Cross-repository integration

---

## Basic Model Tests

**File:** `pkg/models/basic_test.go`

**Purpose:** Simple unit tests validating basic model functionality and field initialization.

### Test Coverage
- Player creation and field access
- GameSession creation with board initialization
- Move validation logic
- PlayerStats win rate calculation

---

## Model Persistence Property Tests

**File:** `pkg/models/persistence_property_test.go`

**Build Tag:** `property`

**Purpose:** Property-based tests ensuring data model integrity and serialization correctness.

### Test Structure

```
TestGameDataPersistence (Property 10)
├── Player data serialization round trip
├── GameSession data integrity
├── PlayerStats calculation accuracy
└── Move validation consistency
```

### Generators
- `genPlayer()` - Generates random Player instances
- `genGameSession()` - Generates random GameSession instances
- `genPlayerStats()` - Generates random PlayerStats instances
- `genMove()` - Generates random Move instances

### Requirements Validated
- **Property 10:** Game Data Persistence (Requirements 6.1, 6.2)

---

## Running Tests

### Run All Unit Tests
```bash
go test ./... -v
```

### Run Property-Based Tests
```bash
go test ./... -v -tags=property
```

### Run Specific Package Tests
```bash
# Game service tests
go test ./internal/game/... -v

# Repository tests
go test ./internal/database/repositories/... -v

# Model tests
go test ./pkg/models/... -v
```

### Run Tests with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Test Conventions

### Naming
- Test functions: `Test<FunctionName>_<Scenario>`
- Sub-tests: Descriptive lowercase with underscores
- Mock types: `Mock<InterfaceName>`

### Structure
- Use `testify/suite` for repository tests (setup/teardown)
- Use `testify/assert` and `testify/require` for assertions
- Use `gopter` for property-based testing

### Database Testing
- In-memory SQLite for repository tests
- Auto-migrate schemas in SetupTest
- Close connections in TearDownTest

### Property Testing
- Minimum 100 iterations per property
- Use build tag `property` for property tests
- Reference design document properties in comments
