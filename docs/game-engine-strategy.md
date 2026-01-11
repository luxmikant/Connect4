# Connect 4 Game Engine Strategy

## Overview

This document defines the strategy and architecture for the Connect 4 game engine. The engine is responsible for managing game state, validating moves, detecting wins/draws, and coordinating game sessions.

## Core Components

### 1. Game Engine Interface

The game engine exposes a clean interface following the Interface Segregation Principle:

```go
type Engine interface {
    // Game lifecycle
    CreateGame(ctx context.Context, player1, player2 string) (*GameSession, error)
    GetGame(ctx context.Context, gameID string) (*GameSession, error)
    
    // Move operations
    MakeMove(ctx context.Context, gameID, playerUsername string, column int) (*MoveResult, error)
    ValidateMove(ctx context.Context, gameID, playerUsername string, column int) error
    
    // Game state checks
    CheckGameEnd(ctx context.Context, game *GameSession) (*GameEndResult, error)
    IsPlayerTurn(ctx context.Context, game *GameSession, playerUsername string) bool
}
```

### 2. Board Representation

The game board uses a 6x7 grid (6 rows, 7 columns) with gravity-based disc placement:

```
Column:  0   1   2   3   4   5   6
       ┌───┬───┬───┬───┬───┬───┬───┐
Row 5  │   │   │   │   │   │   │   │  ← Top
       ├───┼───┼───┼───┼───┼───┼───┤
Row 4  │   │   │   │   │   │   │   │
       ├───┼───┼───┼───┼───┼───┼───┤
Row 3  │   │   │   │   │   │   │   │
       ├───┼───┼───┼───┼───┼───┼───┤
Row 2  │   │   │   │   │   │   │   │
       ├───┼───┼───┼───┼───┼───┼───┤
Row 1  │   │   │   │   │   │   │   │
       ├───┼───┼───┼───┼───┼───┼───┤
Row 0  │   │   │   │   │   │   │   │  ← Bottom
       └───┴───┴───┴───┴───┴───┴───┘
```

**Data Structure:**
```go
type Board struct {
    Grid   [6][7]PlayerColor  // 6 rows × 7 columns
    Height [7]int             // Track column heights for O(1) placement
}
```

## Move Validation Strategy

### Validation Rules

1. **Column Range Check**: Column must be between 0-6
2. **Column Capacity Check**: Column must not be full (height < 6)
3. **Turn Validation**: Must be the player's turn
4. **Game State Check**: Game must be in "in_progress" status

### Validation Flow

```
┌─────────────────┐
│  Receive Move   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Is game active? │──No─▶│  Return Error   │
└────────┬────────┘     └─────────────────┘
         │ Yes
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Is player turn? │──No─▶│  Return Error   │
└────────┬────────┘     └─────────────────┘
         │ Yes
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Is column valid │──No─▶│  Return Error   │
│   (0-6)?        │     └─────────────────┘
└────────┬────────┘
         │ Yes
         ▼
┌─────────────────┐     ┌─────────────────┐
│ Is column full? │─Yes─▶│  Return Error   │
└────────┬────────┘     └─────────────────┘
         │ No
         ▼
┌─────────────────┐
│  Move is Valid  │
└─────────────────┘
```

## Win Detection Strategy

### Detection Algorithm

The engine checks for four-in-a-row in all four directions after each move:

1. **Horizontal**: Check rows for 4 consecutive same-color discs
2. **Vertical**: Check columns for 4 consecutive same-color discs
3. **Diagonal (↘)**: Check top-left to bottom-right diagonals
4. **Diagonal (↙)**: Check top-right to bottom-left diagonals

### Scanning Boundaries

```go
// Horizontal: rows 0-5, columns 0-3 (start positions)
for row := 0; row < 6; row++ {
    for col := 0; col < 4; col++ {
        // Check grid[row][col] through grid[row][col+3]
    }
}

// Vertical: rows 0-2, columns 0-6 (start positions)
for row := 0; row < 3; row++ {
    for col := 0; col < 7; col++ {
        // Check grid[row][col] through grid[row+3][col]
    }
}

// Diagonal (↘): rows 0-2, columns 0-3
for row := 0; row < 3; row++ {
    for col := 0; col < 4; col++ {
        // Check grid[row][col], grid[row+1][col+1], grid[row+2][col+2], grid[row+3][col+3]
    }
}

// Diagonal (↙): rows 0-2, columns 3-6
for row := 0; row < 3; row++ {
    for col := 3; col < 7; col++ {
        // Check grid[row][col], grid[row+1][col-1], grid[row+2][col-2], grid[row+3][col-3]
    }
}
```

### Win Detection Visualization

```
Horizontal Win:          Vertical Win:           Diagonal Wins:
┌───┬───┬───┬───┐       ┌───┐                   ┌───┐           ┌───┐
│ R │ R │ R │ R │       │ R │                   │ R │           │ R │
└───┴───┴───┴───┘       ├───┤                   └───┘           └───┘
                        │ R │                       └───┐   ┌───┘
                        ├───┤                       │ R │   │ R │
                        │ R │                       └───┘   └───┘
                        ├───┤                           └───┐   ┌───┘
                        │ R │                           │ R │   │ R │
                        └───┘                           └───┘   └───┘
                                                            └───┐   ┌───┘
                                                            │ R │   │ R │
                                                            └───┘   └───┘
```

## Draw Detection Strategy

A draw occurs when:
- The board is completely full (all 42 positions occupied)
- No player has achieved four-in-a-row

```go
func (b *Board) IsFull() bool {
    for col := 0; col < 7; col++ {
        if b.Height[col] < 6 {
            return false  // At least one column has space
        }
    }
    return true  // All columns are full
}
```

## Game State Machine

```
                    ┌─────────────┐
                    │   WAITING   │
                    └──────┬──────┘
                           │ Both players joined
                           ▼
                    ┌─────────────┐
              ┌────▶│ IN_PROGRESS │◀────┐
              │     └──────┬──────┘     │
              │            │            │
              │     ┌──────┴──────┐     │
              │     │             │     │
              │     ▼             ▼     │
         No winner         Win detected │
         No draw                        │
              │     ┌─────────────┐     │
              │     │  COMPLETED  │     │
              │     └─────────────┘     │
              │            ▲            │
              │            │            │
              └────────────┴────────────┘
                     Draw or Win

                    ┌─────────────┐
                    │  ABANDONED  │
                    └─────────────┘
                           ▲
                           │ Player disconnect timeout
```

## Turn Management

### Turn Alternation

- Player 1 (Red) always starts first
- Turns alternate after each valid move
- Current turn is tracked in the GameSession

```go
// After a valid move
if game.CurrentTurn == PlayerColorRed {
    game.CurrentTurn = PlayerColorYellow
} else {
    game.CurrentTurn = PlayerColorRed
}
```

### Player Color Assignment

```go
func (gs *GameSession) GetPlayerColor(username string) PlayerColor {
    if username == gs.Player1 {
        return PlayerColorRed    // Player 1 is always Red
    }
    return PlayerColorYellow     // Player 2 is always Yellow
}
```

## Move Processing Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                        MakeMove Flow                              │
└──────────────────────────────────────────────────────────────────┘

1. Get current game state
         │
         ▼
2. Validate move (column, turn, game status)
         │
         ▼
3. Calculate landing row (board.Height[column])
         │
         ▼
4. Place disc on board (grid[row][column] = player)
         │
         ▼
5. Increment column height (board.Height[column]++)
         │
         ▼
6. Record move in database
         │
         ▼
7. Switch turns
         │
         ▼
8. Check for win/draw
         │
         ├─── Win detected ──▶ Update status to COMPLETED, set winner
         │
         ├─── Draw detected ─▶ Update status to COMPLETED, no winner
         │
         └─── Game continues ─▶ Return updated state
```

## Error Handling Strategy

### Error Types

| Error | Condition | HTTP Status |
|-------|-----------|-------------|
| `ErrInvalidMove` | Column out of range or full | 400 |
| `ErrGameNotFound` | Game ID doesn't exist | 404 |
| `ErrNotPlayerTurn` | Wrong player attempting move | 403 |
| `ErrGameEnded` | Move attempted on completed game | 400 |
| `ErrPlayerNotFound` | Player not in game | 404 |

### Error Response Format

```go
type GameError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

## Performance Considerations

### O(1) Operations

- **Column height lookup**: `board.Height[column]`
- **Move placement**: Direct array access
- **Turn check**: Simple string comparison

### O(n) Operations

- **Win detection**: Scans fixed number of positions (69 checks max)
- **Draw detection**: Checks 7 column heights

### Memory Efficiency

- Board uses fixed-size arrays (no dynamic allocation)
- Height array eliminates need to scan columns for placement
- Game state stored in single struct

## Testing Strategy

### Property-Based Tests

1. **Move Validation Properties**
   - Valid columns (0-6) accepted for non-full columns
   - Invalid columns rejected
   - Discs placed in lowest available position
   - Turn enforcement

2. **Win Detection Properties**
   - Horizontal 4-in-a-row detected
   - Vertical 4-in-a-row detected
   - Diagonal 4-in-a-row detected (both directions)
   - No false positives on non-winning boards

3. **Draw Detection Properties**
   - Full boards without winner detected as draw
   - Incomplete boards not detected as draw

### Unit Tests

- Specific win scenarios
- Edge cases (corner wins, boundary conditions)
- Error condition handling
- Game state transitions

## Integration Points

### Repository Layer

```go
type Engine struct {
    gameRepo repositories.GameSessionRepository
    moveRepo repositories.MoveRepository
}
```

### WebSocket Events

The engine triggers events for real-time updates:
- `move_made`: After successful move
- `game_ended`: When win or draw detected
- `turn_changed`: After turn switch

### Analytics Events

Events published to Kafka:
- `game_started`: New game created
- `move_made`: Each move with timing data
- `game_completed`: Final game state and outcome
