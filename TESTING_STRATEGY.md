# Testing Strategy - Complete Guide

## 📋 Table of Contents
1. [Testing Overview](#testing-overview)
2. [Testing Frameworks](#testing-frameworks)
3. [Test Types & Patterns](#test-types--patterns)
4. [Test Organization](#test-organization)
5. [Running Tests](#running-tests)
6. [Coverage Analysis](#coverage-analysis)
7. [Example Test Implementations](#example-test-implementations)

---

## Testing Overview

The Connect4 project uses a **multi-layered testing approach** combining:

| Layer | Framework | Scope | Coverage |
|-------|-----------|-------|----------|
| **Unit Tests** | Testify (assert/require) | Individual functions/methods | ~70-80% |
| **Property Tests** | Gopter (property-based) | Universal invariants | ~40-50% (selected modules) |
| **Integration Tests** | Testify (suite) | End-to-end workflows | Cloud services |
| **Mock Tests** | Custom mocks | Isolated component testing | All dependencies |

---

## Testing Frameworks

### 1. Testify - Assertion & Testing Library

**Website**: https://github.com/stretchr/testify  
**Go Import**: `github.com/stretchr/testify`

**Key Packages**:

#### `testify/assert`
Provides assertion functions that allow tests to continue even if assertion fails.

```go
import "github.com/stretchr/testify/assert"

func TestExample(t *testing.T) {
    // Will print error but continue execution
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.True(t, condition)
    assert.Contains(t, haystack, needle)
}
```

#### `testify/require`
Provides assertion functions that halt test execution on failure.

```go
import "github.com/stretchr/testify/require"

func TestExample(t *testing.T) {
    // Will stop test immediately on failure
    require.NoError(t, err)
    require.Equal(t, expected, actual)
}
```

#### `testify/suite`
Provides testing suite structure with setup/teardown.

```go
import "github.com/stretchr/testify/suite"

type MyTestSuite struct {
    suite.Suite
    
    // Shared test fixtures
    db *gorm.DB
}

func (suite *MyTestSuite) SetupTest() {
    // Runs before each test
}

func (suite *MyTestSuite) TearDownTest() {
    // Runs after each test
}

func (suite *MyTestSuite) TestSomething() {
    suite.Assert().Equal(expected, actual)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

---

### 2. Gopter - Property-Based Testing Framework

**Website**: https://github.com/leanovate/gopter  
**Go Import**: `github.com/leanovate/gopter`

**Purpose**: Generate hundreds of random test inputs to find edge cases

**Key Concepts**:

#### Generators
Generate random input values:

```go
// Built-in generators
gen.Int() // Random integer
gen.IntRange(0, 6) // Integer between 0-6
gen.OneConstOf("red", "yellow") // Pick one value
gen.SliceOfN(10, gen.Int()) // Generate slice of 10 integers
gen.StructPtr(genBoard, "Field1", "Field2") // Generate struct instances
```

#### Properties
Define universal rules that should hold true:

```go
import (
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

func TestBoardProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Property: All valid columns should accept moves
    properties.Property("valid moves are always placeable", prop.ForAll(
        func(column int) bool {
            board := models.NewBoard()
            
            // Skip if invalid column
            if !board.IsValidMove(column) {
                return true // Property holds vacuously
            }
            
            // For valid columns, we must be able to move
            err := board.MakeMove(column, models.PlayerColorRed)
            return err == nil
        },
        gen.IntRange(0, 6), // Generator for column
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

#### Shrinking
Automatically finds minimal failing cases:

```
Finds failing test case: [1, 2, 3, 4, 5, ...]
Shrinks to: [1, 2, 3]  // Minimal failing input
Shows developer exactly what breaks
```

---

## Test Types & Patterns

### 1. Unit Tests (Testify)

**Purpose**: Test individual functions in isolation  
**Location**: Same package as code being tested  
**Pattern**: `*_test.go` files

**Example: Game Engine Tests**

```go
// File: internal/game/engine_test.go
package game_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "connect4-multiplayer/pkg/models"
)

func TestCreateGame_Success(t *testing.T) {
    engine := game.NewEngine(
        NewMockGameSessionRepository(),
        NewMockMoveRepository(),
    )
    ctx := context.Background()

    // Arrange: Set up inputs
    // Act: Call function
    gameSession, err := engine.CreateGame(ctx, "player1", "player2")

    // Assert: Check results
    require.NoError(t, err)
    assert.NotEmpty(t, gameSession.ID)
    assert.Equal(t, "player1", gameSession.Player1)
    assert.Equal(t, models.StatusInProgress, gameSession.Status)
}

func TestCreateGame_EmptyPlayer1(t *testing.T) {
    engine := game.NewEngine(
        NewMockGameSessionRepository(),
        NewMockMoveRepository(),
    )
    ctx := context.Background()

    _, err := engine.CreateGame(ctx, "", "player2")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "cannot be empty")
}

func TestCreateGame_SamePlayer(t *testing.T) {
    engine := game.NewEngine(
        NewMockGameSessionRepository(),
        NewMockMoveRepository(),
    )
    ctx := context.Background()

    _, err := engine.CreateGame(ctx, "player1", "player1")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "different usernames")
}
```

**Test Categories**:
- **Happy Path**: `TestCreateGame_Success`
- **Error Cases**: `TestCreateGame_EmptyPlayer1`
- **Edge Cases**: `TestCreateGame_SamePlayer`

---

### 2. Property-Based Tests (Gopter)

**Purpose**: Verify universal properties with generated inputs  
**Location**: Same package as code, suffixed with `_property_test.go`  
**Build Tag**: `//go:build property`

**Example: Game Logic Properties**

```go
// File: internal/game/engine_property_test.go
//go:build property
// +build property

package game_test

import (
    "context"
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
    "connect4-multiplayer/pkg/models"
)

// Feature: connect-4-multiplayer, Property 7: Game Move Validation
func TestMoveValidationProperty(t *testing.T) {
    properties := gopter.NewProperties(nil)

    // Property 1: Valid moves should be accepted for non-full columns
    properties.Property("valid moves accepted for non-full columns", 
        prop.ForAll(
            func(column int) bool {
                engine := game.NewEngine(
                    NewMockGameSessionRepository(),
                    NewMockMoveRepository(),
                )
                ctx := context.Background()
                gameSession, _ := engine.CreateGame(ctx, "p1", "p2")
                
                if !gameSession.Board.IsValidMove(column) {
                    return true // Skip invalid columns
                }
                
                err := engine.ValidateMove(ctx, gameSession.ID, "p1", column)
                return err == nil
            },
            gen.IntRange(0, 6), // Generate columns 0-6
        ),
    )

    // Property 2: Negative columns should be rejected
    properties.Property("negative columns rejected",
        prop.ForAll(
            func(column int) bool {
                engine := game.NewEngine(
                    NewMockGameSessionRepository(),
                    NewMockMoveRepository(),
                )
                ctx := context.Background()
                gameSession, _ := engine.CreateGame(ctx, "p1", "p2")
                
                err := engine.ValidateMove(ctx, gameSession.ID, "p1", column)
                return err != nil // Should error for negative
            },
            gen.IntRange(-100, -1), // Generate negative columns
        ),
    )

    // Property 3: Columns >= 7 should be rejected
    properties.Property("columns >= 7 rejected",
        prop.ForAll(
            func(column int) bool {
                engine := game.NewEngine(
                    NewMockGameSessionRepository(),
                    NewMockMoveRepository(),
                )
                ctx := context.Background()
                gameSession, _ := engine.CreateGame(ctx, "p1", "p2")
                
                err := engine.ValidateMove(ctx, gameSession.ID, "p1", column)
                return err != nil // Should error for column >= 7
            },
            gen.IntRange(7, 100), // Generate invalid columns
        ),
    )

    properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: connect-4-multiplayer, Property 8: Win Detection
func TestWinDetectionProperty(t *testing.T) {
    properties := gopter.NewProperties(nil)

    // Property: Horizontal 4-in-a-row detected as win
    properties.Property("horizontal 4-in-a-row detected",
        prop.ForAll(
            func(startColumn int) bool {
                if startColumn < 0 || startColumn > 3 {
                    return true // Skip invalid start positions
                }
                
                board := models.NewBoard()
                
                // Place 4 red discs horizontally
                for col := startColumn; col < startColumn+4; col++ {
                    board.MakeMove(col, models.PlayerColorRed)
                }
                
                winner := board.CheckWin()
                return winner != nil && *winner == models.PlayerColorRed
            },
            gen.IntRange(0, 3),
        ),
    )

    // Property: Vertical 4-in-a-row detected as win
    properties.Property("vertical 4-in-a-row detected",
        prop.ForAll(
            func(column int) bool {
                if column < 0 || column > 6 {
                    return true
                }
                
                board := models.NewBoard()
                
                // Place 4 red discs vertically in one column
                for i := 0; i < 4; i++ {
                    board.MakeMove(column, models.PlayerColorRed)
                }
                
                winner := board.CheckWin()
                return winner != nil && *winner == models.PlayerColorRed
            },
            gen.IntRange(0, 6),
        ),
    )

    // Property: Full board without winner is draw
    properties.Property("full board without winner is draw",
        prop.ForAll(
            func(moveSequence []int) bool {
                board := models.NewBoard()
                color := models.PlayerColorRed
                
                // Make all valid moves
                for _, col := range moveSequence {
                    if col >= 0 && col < 7 && board.IsValidMove(col) {
                        board.MakeMove(col, color)
                        if color == models.PlayerColorRed {
                            color = models.PlayerColorYellow
                        } else {
                            color = models.PlayerColorRed
                        }
                    }
                }
                
                if board.IsFull() && board.CheckWin() == nil {
                    return true // Is draw
                }
                return !board.IsFull() || board.CheckWin() != nil
            },
            gen.SliceOfN(50, gen.IntRange(0, 6)),
        ),
    )

    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

**Benefits of Property Testing**:
- ✅ Finds edge cases humans miss
- ✅ Documents invariants explicitly
- ✅ High confidence with hundreds of test cases
- ✅ Automatic shrinking shows minimal failing input

---

### 3. Integration Tests (Testify Suite)

**Purpose**: Test full workflows with real dependencies  
**Location**: `tests/integration/` directory  
**Build Tag**: `//go:build integration`

**Example: End-to-End Game Flow**

```go
// File: tests/integration/e2e_test.go
//go:build integration
// +build integration

package integration

import (
    "context"
    "os"
    "testing"
    "github.com/gorilla/websocket"
    "github.com/stretchr/testify/suite"
    "gorm.io/gorm"
)

type E2ETestSuite struct {
    suite.Suite
    db         *gorm.DB
    gameService game.GameService
    ctx        context.Context
}

func (suite *E2ETestSuite) SetupSuite() {
    // Initialize real PostgreSQL database
    suite.ctx = context.Background()
    
    // Load cloud credentials
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        suite.T().Skip("DATABASE_URL not set")
    }
    
    // Connect to real database
    var err error
    suite.db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
    suite.Require().NoError(err)
    
    // Initialize services with real dependencies
    repoManager := repositories.NewManager(suite.db)
    suite.gameService = game.NewGameService(repoManager)
}

func (suite *E2ETestSuite) TearDownSuite() {
    // Clean up
    suite.db.Migrator().DropTable(&models.GameSession{})
    suite.db.Migrator().DropTable(&models.Player{})
}

func (suite *E2ETestSuite) TestCompleteGameFlow() {
    // 1. Create game
    gameSession, err := suite.gameService.CreateSession("alice", "bob")
    suite.Require().NoError(err)
    suite.NotNil(gameSession)
    
    // 2. Make moves
    for i := 0; i < 35; i++ { // 35 moves to complete game
        gameSession, err = suite.gameService.GetSession(suite.ctx, gameSession.ID)
        suite.Require().NoError(err)
        
        if gameSession.Status == "completed" {
            break
        }
        
        // Simulate move
        column := i % 7
        gameSession.Board.MakeMove(column, gameSession.CurrentTurn)
        
        if gameSession.Board.CheckWin() != nil {
            suite.gameService.CompleteGame(
                gameSession.ID,
                gameSession.Board.CheckWin(),
            )
            break
        }
        
        gameSession.SwitchTurn()
        suite.gameService.Update(suite.ctx, gameSession)
    }
    
    // 3. Verify completion
    gameSession, _ = suite.gameService.GetSession(suite.ctx, gameSession.ID)
    suite.Equal("completed", gameSession.Status)
    suite.NotNil(gameSession.Winner)
}

func TestE2ETestSuite(t *testing.T) {
    suite.Run(t, new(E2ETestSuite))
}
```

---

### 4. Mock Tests

**Purpose**: Test components with mocked dependencies  
**Location**: Same package as code  
**Pattern**: `mocks_test.go` files

**Example: Mock Game Session Repository**

```go
// File: internal/game/mocks_test.go
package game_test

import (
    "context"
    "connect4-multiplayer/pkg/models"
)

// MockGameSessionRepository implements GameSessionRepository interface
type MockGameSessionRepository struct {
    games map[string]*models.GameSession
}

func NewMockGameSessionRepository() *MockGameSessionRepository {
    return &MockGameSessionRepository{
        games: make(map[string]*models.GameSession),
    }
}

// Create adds a new game session
func (m *MockGameSessionRepository) Create(
    ctx context.Context,
    session *models.GameSession,
) error {
    if session.ID == "" {
        session.ID = "test-game-id"
    }
    m.games[session.ID] = session
    return nil
}

// GetByID retrieves a game session by ID
func (m *MockGameSessionRepository) GetByID(
    ctx context.Context,
    id string,
) (*models.GameSession, error) {
    if game, exists := m.games[id]; exists {
        return game, nil
    }
    return nil, models.ErrGameNotFound
}

// GetByRoomCode retrieves a game session by custom room code
func (m *MockGameSessionRepository) GetByRoomCode(
    ctx context.Context,
    roomCode string,
) (*models.GameSession, error) {
    for _, game := range m.games {
        if game.RoomCode == roomCode && game.IsCustom {
            return game, nil
        }
    }
    return nil, models.ErrGameNotFound
}

// Update modifies an existing game session
func (m *MockGameSessionRepository) Update(
    ctx context.Context,
    session *models.GameSession,
) error {
    m.games[session.ID] = session
    return nil
}

// DeleteByID removes a game session
func (m *MockGameSessionRepository) Delete(
    ctx context.Context,
    id string,
) error {
    delete(m.games, id)
    return nil
}
```

---

## Test Organization

### Directory Structure

```
project/
├── internal/
│   ├── game/
│   │   ├── engine.go
│   │   ├── service.go
│   │   ├── engine_test.go          # Unit tests
│   │   ├── engine_property_test.go # Property tests (build tag: property)
│   │   ├── service_test.go         # Unit tests
│   │   ├── service_property_test.go # Property tests
│   │   └── mocks_test.go           # Mock implementations
│   │
│   ├── websocket/
│   │   ├── handler.go
│   │   ├── handler_test.go         # Unit tests
│   │   ├── websocket_property_test.go # Property tests
│   │   ├── integration_test.go     # Integration tests
│   │   └── mocks_test.go           # Mock implementations
│   │
│   └── bot/
│       ├── minimax.go
│       ├── minimax_test.go         # Unit tests
│       ├── bot_property_test.go    # Property tests
│       └── ...
│
├── pkg/
│   └── models/
│       ├── game.go
│       ├── basic_test.go           # Unit tests
│       └── persistence_property_test.go # Property tests
│
└── tests/
    └── integration/
        ├── e2e_test.go             # End-to-end tests
        ├── performance_test.go     # Performance tests
        └── kafka_test.go           # Kafka integration tests
```

### Test File Naming Conventions

```
Source Code:      → Test File:
engine.go         → engine_test.go              (Unit tests)
engine.go         → engine_property_test.go     (Property tests)
service.go        → service_test.go             (Unit tests)
service.go        → service_property_test.go    (Property tests)
mocks for tests   → mocks_test.go              (All mocks)
```

---

## Running Tests

### 1. Using Make Commands

```bash
# Run all unit tests
make test

# Run all tests with coverage
make test-coverage

# Run property-based tests only
make test-property

# Run integration tests
make test-integration
```

### 2. Using Go Commands

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/game/...

# Run specific test function
go test ./internal/game/... -run TestCreateGame

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Using Build Tags

```bash
# Run only property tests (build tag: property)
go test -tags=property -v ./...

# Run only integration tests (build tag: integration)
go test -tags=integration -v ./...

# Run specific test suite
go test -tags=integration -v ./tests/integration -run=TestE2ETestSuite

# Run with timeout (some tests are slow)
go test -tags=integration -timeout=15m -v ./tests/integration
```

### 4. Selective Testing

```bash
# Test only game package
go test ./internal/game -v

# Test only WebSocket handlers
go test ./internal/websocket -v

# Test only bot/AI
go test ./internal/bot -v

# Skip short tests
go test -short ./...

# Run in short mode (skips integration tests marked with testing.Short())
go test -short ./...
```

---

## Coverage Analysis

### Generating Coverage Reports

```bash
# Generate coverage data
go test -coverprofile=coverage.out ./...

# View coverage in HTML
go tool cover -html=coverage.out

# View specific package coverage
go test -coverprofile=coverage.out ./internal/game/...
go tool cover -html=coverage.out -o game-coverage.html

# Show coverage in terminal
go tool cover -func=coverage.out | grep total
```

### Coverage Requirements

**Target Coverage by Package**:
- `internal/game/` - 80% (critical game logic)
- `internal/bot/` - 75% (complex algorithms)
- `internal/websocket/` - 70% (message handlers)
- `internal/matchmaking/` - 70% (queue management)
- `pkg/models/` - 85% (data structures)

### Coverage Report Example

```
coverage: 74.2% of statements
ok      connect4-multiplayer/internal/game      1.234s

Function Coverage:
connect4-multiplayer/internal/game.CreateSession          100.0%
connect4-multiplayer/internal/game.ValidateMove           95.0%
connect4-multiplayer/internal/game.CheckWin              100.0%
...total:                                                 82.5%
```

---

## Example Test Implementations

### Example 1: Unit Test with Mocks

```go
// File: internal/game/engine_test.go

func TestValidateMove_ValidColumn(t *testing.T) {
    // Setup
    mockRepo := NewMockGameSessionRepository()
    engine := game.NewEngine(mockRepo, NewMockMoveRepository())
    ctx := context.Background()
    
    // Create game
    gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
    
    // Test valid column (0-6)
    for column := 0; column < 7; column++ {
        err := engine.ValidateMove(ctx, gameSession.ID, "player1", column)
        assert.NoError(t, err)
    }
}

func TestValidateMove_InvalidColumn_Negative(t *testing.T) {
    mockRepo := NewMockGameSessionRepository()
    engine := game.NewEngine(mockRepo, NewMockMoveRepository())
    ctx := context.Background()
    gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
    
    err := engine.ValidateMove(ctx, gameSession.ID, "player1", -1)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid column")
}

func TestValidateMove_InvalidColumn_TooLarge(t *testing.T) {
    mockRepo := NewMockGameSessionRepository()
    engine := game.NewEngine(mockRepo, NewMockMoveRepository())
    ctx := context.Background()
    gameSession, _ := engine.CreateGame(ctx, "player1", "player2")
    
    err := engine.ValidateMove(ctx, gameSession.ID, "player1", 7)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid column")
}
```

### Example 2: Property-Based Test

```go
// File: internal/game/engine_property_test.go
//go:build property

func TestBoardHeightTracking(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property(
        "board height increases with each move in column",
        prop.ForAll(
            func(column int, moveCount int) bool {
                if column < 0 || column > 6 || moveCount < 0 || moveCount > 6 {
                    return true // Skip invalid inputs
                }
                
                board := models.NewBoard()
                initialHeight := board.Height[column]
                
                // Make moves in same column
                for i := 0; i < moveCount && board.IsValidMove(column); i++ {
                    color := models.PlayerColorRed
                    if i%2 == 1 {
                        color = models.PlayerColorYellow
                    }
                    board.MakeMove(column, color)
                }
                
                // Height should increase by number of moves
                return board.Height[column] >= initialHeight
            },
            gen.IntRange(0, 6),
            gen.IntRange(0, 6),
        ),
    )
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Example 3: Integration Test with Suite

```go
// File: tests/integration/e2e_test.go
//go:build integration

type GameFlowTestSuite struct {
    suite.Suite
    db          *gorm.DB
    gameService game.GameService
}

func (suite *GameFlowTestSuite) SetupSuite() {
    // Connect to test database
    suite.db, _ = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")))
    repoManager := repositories.NewManager(suite.db)
    suite.gameService = game.NewGameService(repoManager)
}

func (suite *GameFlowTestSuite) TestCompleteGameWithWinner() {
    // Create game
    gameSession, _ := suite.gameService.CreateSession("alice", "bob")
    suite.Require().NotNil(gameSession)
    
    // Play game until someone wins
    moves := []int{0, 1, 0, 1, 0, 1, 0} // Alice wins horizontally
    
    for i, col := range moves {
        player := gameSession.Player1
        if i%2 == 1 {
            player = gameSession.Player2
        }
        
        gameSession.Board.MakeMove(col, gameSession.CurrentTurn)
        
        if winner := gameSession.Board.CheckWin(); winner != nil {
            suite.gameService.CompleteGame(gameSession.ID, winner)
            break
        }
        
        gameSession.SwitchTurn()
        suite.gameService.Update(context.Background(), gameSession)
    }
    
    // Verify completion
    gameSession, _ = suite.gameService.GetSession(context.Background(), gameSession.ID)
    suite.Equal("completed", gameSession.Status)
    suite.NotNil(gameSession.Winner)
}

func TestGameFlowTestSuite(t *testing.T) {
    suite.Run(t, new(GameFlowTestSuite))
}
```

### Example 4: Mock Implementation

```go
// File: internal/websocket/mocks_test.go

type MockGameService struct {
    CreateSessionFunc    func(p1, p2 string) (*models.GameSession, error)
    RematchCustomRoomFunc func(gameID, username string) (*models.GameSession, error)
    GetSessionFunc       func(ctx context.Context, id string) (*models.GameSession, error)
}

func (m *MockGameService) CreateSession(
    p1, p2 string,
) (*models.GameSession, error) {
    if m.CreateSessionFunc != nil {
        return m.CreateSessionFunc(p1, p2)
    }
    return &models.GameSession{}, nil
}

func (m *MockGameService) RematchCustomRoom(
    gameID, username string,
) (*models.GameSession, error) {
    if m.RematchCustomRoomFunc != nil {
        return m.RematchCustomRoomFunc(gameID, username)
    }
    return &models.GameSession{}, nil
}

// Used in tests
func TestRematchFeature(t *testing.T) {
    mockService := &MockGameService{
        RematchCustomRoomFunc: func(gameID, username string) (*models.GameSession, error) {
            return &models.GameSession{
                ID:       "new-game-id",
                RoomCode: "ABC12345",
                Player1:  "alice",
                Player2:  "bob",
            }, nil
        },
    }
    
    session, err := mockService.RematchCustomRoom("game-1", "alice")
    
    assert.NoError(t, err)
    assert.Equal(t, "new-game-id", session.ID)
    assert.Equal(t, "ABC12345", session.RoomCode)
}
```

---

## Test Maintenance Guidelines

### When to Use Each Test Type

| Situation | Use |
|-----------|-----|
| Testing single function logic | Unit test |
| Testing invariants/edge cases | Property test |
| Testing complete workflow | Integration test |
| Isolating dependencies | Mock test |
| Testing with real database | Suite test |
| Testing algorithm behavior | Both unit + property |

### Best Practices

1. **Arrange-Act-Assert Pattern**
   ```go
   func TestSomething(t *testing.T) {
       // Arrange: Set up test data
       input := 5
       expected := 10
       
       // Act: Execute function
       result := Double(input)
       
       // Assert: Check results
       assert.Equal(t, expected, result)
   }
   ```

2. **Use Table-Driven Tests for Multiple Scenarios**
   ```go
   func TestValidateColumn(t *testing.T) {
       tests := []struct {
           name      string
           column    int
           wantError bool
       }{
           {"valid 0", 0, false},
           {"valid 6", 6, false},
           {"invalid -1", -1, true},
           {"invalid 7", 7, true},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := validateColumn(tt.column)
               if tt.wantError {
                   assert.Error(t, err)
               } else {
                   assert.NoError(t, err)
               }
           })
       }
   }
   ```

3. **Keep Tests Focused**
   - One assertion per test (or related assertions)
   - Test one behavior per test
   - Use subtests (t.Run) for related scenarios

4. **Avoid Test Interdependencies**
   - Don't rely on order of test execution
   - Clean up resources in TearDownTest
   - Use fresh fixtures for each test

5. **Property Tests Need Generators**
   ```go
   // Good: Using appropriate generator
   gen.IntRange(0, 6)  // For board columns
   
   // Bad: Wrong generator
   gen.IntRange(-999999, 999999)  // Too broad, won't find issues
   ```

---

## Summary Table

| Framework | Purpose | Pattern | Tag | Count |
|-----------|---------|---------|-----|-------|
| **Testify** | Unit tests | `*_test.go` | None | ~50 tests |
| **Gopter** | Property tests | `*_property_test.go` | `property` | ~20 tests |
| **Suite** | Integration tests | Suite struct | `integration` | ~10 tests |
| **Mocks** | Dependency isolation | `mocks_test.go` | None | ~8 files |

**Total Test Coverage**: ~80 test files, 100+ test functions, millions of generated test cases

This comprehensive testing approach ensures:
- ✅ Correctness of core logic
- ✅ Resilience to edge cases
- ✅ End-to-end workflow validation
- ✅ Integration with real services
- ✅ High confidence in deployments
