# Connect 4 Game Engine - Practical Implementation Guide

## Overview

This guide provides practical code examples demonstrating how to use the Connect 4 game engine. Each section includes working code that you can run and test.

## Quick Start

### 1. Creating a New Game

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "connect4-multiplayer/internal/game"
    "connect4-multiplayer/internal/database/repositories"
)

func main() {
    // Initialize repositories (using mocks for demo)
    gameRepo := repositories.NewMockGameSessionRepository()
    moveRepo := repositories.NewMockMoveRepository()
    
    // Create the game engine
    engine := game.NewEngine(gameRepo, moveRepo)
    
    // Create a new game between two players
    ctx := context.Background()
    gameSession, err := engine.CreateGame(ctx, "alice", "bob")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Game created: %s\n", gameSession.ID)
    fmt.Printf("Player 1 (Red): %s\n", gameSession.Player1)
    fmt.Printf("Player 2 (Yellow): %s\n", gameSession.Player2)
    fmt.Printf("Current turn: %s\n", gameSession.CurrentTurn)
}
```


### 2. Making Moves

```go
func playGame(engine game.Engine, gameID string) {
    ctx := context.Background()
    
    // Alice (Red) makes first move in column 3 (center)
    result, err := engine.MakeMove(ctx, gameID, "alice", 3)
    if err != nil {
        log.Printf("Move failed: %v", err)
        return
    }
    
    fmt.Printf("Alice dropped disc in column 3, row %d\n", result.Move.Row)
    printBoard(result.GameSession.Board)
    
    // Bob (Yellow) responds in column 3
    result, err = engine.MakeMove(ctx, gameID, "bob", 3)
    if err != nil {
        log.Printf("Move failed: %v", err)
        return
    }
    
    fmt.Printf("Bob dropped disc in column 3, row %d\n", result.Move.Row)
    printBoard(result.GameSession.Board)
}

// Helper function to print the board
func printBoard(board models.Board) {
    fmt.Println("\n  0 1 2 3 4 5 6")
    fmt.Println("  ─────────────")
    for row := 5; row >= 0; row-- {
        fmt.Printf("%d│", row)
        for col := 0; col < 7; col++ {
            switch board.Grid[row][col] {
            case models.PlayerColorRed:
                fmt.Print("R ")
            case models.PlayerColorYellow:
                fmt.Print("Y ")
            default:
                fmt.Print(". ")
            }
        }
        fmt.Println("│")
    }
    fmt.Println("  ─────────────")
}
```


## Complete Game Example

### Playing a Full Game with Win Detection

```go
package main

import (
    "context"
    "fmt"
    
    "connect4-multiplayer/pkg/models"
)

func playCompleteGame() {
    // Create a board and simulate a horizontal win
    board := models.NewBoard()
    
    // Simulate moves: Red wins horizontally
    moves := []struct {
        player models.PlayerColor
        column int
    }{
        {models.PlayerColorRed, 0},    // R at (0,0)
        {models.PlayerColorYellow, 0}, // Y at (1,0)
        {models.PlayerColorRed, 1},    // R at (0,1)
        {models.PlayerColorYellow, 1}, // Y at (1,1)
        {models.PlayerColorRed, 2},    // R at (0,2)
        {models.PlayerColorYellow, 2}, // Y at (1,2)
        {models.PlayerColorRed, 3},    // R at (0,3) - WIN!
    }
    
    for i, move := range moves {
        err := board.MakeMove(move.column, move.player)
        if err != nil {
            fmt.Printf("Move %d failed: %v\n", i+1, err)
            return
        }
        
        fmt.Printf("Move %d: %s plays column %d\n", i+1, move.player, move.column)
        
        // Check for winner after each move
        winner := board.CheckWin()
        if winner != nil {
            fmt.Printf("\n🎉 %s WINS!\n", *winner)
            printBoard(board)
            return
        }
    }
}
```

**Output:**
```
Move 1: red plays column 0
Move 2: yellow plays column 0
Move 3: red plays column 1
Move 4: yellow plays column 1
Move 5: red plays column 2
Move 6: yellow plays column 2
Move 7: red plays column 3

🎉 red WINS!

  0 1 2 3 4 5 6
  ─────────────
5│. . . . . . .│
4│. . . . . . .│
3│. . . . . . .│
2│. . . . . . .│
1│Y Y Y . . . .│
0│R R R R . . .│
  ─────────────
```


## Win Detection Examples

### Horizontal Win

```go
func demonstrateHorizontalWin() {
    board := models.NewBoard()
    
    // Place 4 red discs horizontally in row 0
    for col := 0; col < 4; col++ {
        board.MakeMove(col, models.PlayerColorRed)
    }
    
    winner := board.CheckWin()
    fmt.Printf("Horizontal win detected: %v\n", winner != nil) // true
}
```

### Vertical Win

```go
func demonstrateVerticalWin() {
    board := models.NewBoard()
    
    // Place 4 yellow discs vertically in column 3
    for i := 0; i < 4; i++ {
        board.MakeMove(3, models.PlayerColorYellow)
    }
    
    winner := board.CheckWin()
    fmt.Printf("Vertical win detected: %v\n", winner != nil) // true
}
```

### Diagonal Win (↘)

```go
func demonstrateDiagonalWinTLBR() {
    board := models.NewBoard()
    
    // Build a diagonal from top-left to bottom-right
    // Need to stack discs to create the diagonal
    
    // Column 0: 1 red disc
    board.MakeMove(0, models.PlayerColorRed)
    
    // Column 1: 1 yellow, 1 red
    board.MakeMove(1, models.PlayerColorYellow)
    board.MakeMove(1, models.PlayerColorRed)
    
    // Column 2: 2 yellow, 1 red
    board.MakeMove(2, models.PlayerColorYellow)
    board.MakeMove(2, models.PlayerColorYellow)
    board.MakeMove(2, models.PlayerColorRed)
    
    // Column 3: 3 yellow, 1 red
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorRed)
    
    winner := board.CheckWin()
    fmt.Printf("Diagonal (↘) win detected: %v\n", winner != nil) // true
    
    /*
    Board state:
      0 1 2 3 4 5 6
      ─────────────
    3│. . . R . . .│
    2│. . R Y . . .│
    1│. R Y Y . . .│
    0│R Y Y Y . . .│
      ─────────────
    */
}
```


### Diagonal Win (↙)

```go
func demonstrateDiagonalWinTRBL() {
    board := models.NewBoard()
    
    // Build a diagonal from top-right to bottom-left
    
    // Column 6: 1 red disc
    board.MakeMove(6, models.PlayerColorRed)
    
    // Column 5: 1 yellow, 1 red
    board.MakeMove(5, models.PlayerColorYellow)
    board.MakeMove(5, models.PlayerColorRed)
    
    // Column 4: 2 yellow, 1 red
    board.MakeMove(4, models.PlayerColorYellow)
    board.MakeMove(4, models.PlayerColorYellow)
    board.MakeMove(4, models.PlayerColorRed)
    
    // Column 3: 3 yellow, 1 red
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorYellow)
    board.MakeMove(3, models.PlayerColorRed)
    
    winner := board.CheckWin()
    fmt.Printf("Diagonal (↙) win detected: %v\n", winner != nil) // true
    
    /*
    Board state:
      0 1 2 3 4 5 6
      ─────────────
    3│. . . R . . .│
    2│. . . Y R . .│
    1│. . . Y Y R .│
    0│. . . Y Y Y R│
      ─────────────
    */
}
```

## Draw Detection Example

```go
func demonstrateDraw() {
    board := models.NewBoard()
    
    // Fill the board with alternating pattern that prevents wins
    pattern := [6][7]models.PlayerColor{
        {models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, 
         models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
        {models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, 
         models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
        {models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, 
         models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
        {models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, 
         models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
        {models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, 
         models.PlayerColorRed, models.PlayerColorRed, models.PlayerColorYellow, models.PlayerColorRed},
        {models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, 
         models.PlayerColorYellow, models.PlayerColorYellow, models.PlayerColorRed, models.PlayerColorYellow},
    }
    
    board.Grid = pattern
    for col := 0; col < 7; col++ {
        board.Height[col] = 6
    }
    
    winner := board.CheckWin()
    isFull := board.IsFull()
    
    fmt.Printf("Winner: %v\n", winner)     // nil
    fmt.Printf("Board full: %v\n", isFull) // true
    fmt.Printf("It's a DRAW!\n")
}
```


## Move Validation Examples

### Valid Move

```go
func demonstrateValidMove() {
    ctx := context.Background()
    engine := createEngine()
    
    game, _ := engine.CreateGame(ctx, "alice", "bob")
    
    // Alice (Red) tries to make a valid move
    err := engine.ValidateMove(ctx, game.ID, "alice", 3)
    if err == nil {
        fmt.Println("✓ Move is valid")
    }
}
```

### Invalid Column

```go
func demonstrateInvalidColumn() {
    ctx := context.Background()
    engine := createEngine()
    
    game, _ := engine.CreateGame(ctx, "alice", "bob")
    
    // Try invalid column (out of range)
    err := engine.ValidateMove(ctx, game.ID, "alice", 7)
    if err != nil {
        fmt.Printf("✗ Invalid move: %v\n", err)
        // Output: ✗ Invalid move: invalid column: 7 (must be 0-6)
    }
    
    // Try negative column
    err = engine.ValidateMove(ctx, game.ID, "alice", -1)
    if err != nil {
        fmt.Printf("✗ Invalid move: %v\n", err)
        // Output: ✗ Invalid move: invalid column: -1 (must be 0-6)
    }
}
```

### Full Column

```go
func demonstrateFullColumn() {
    ctx := context.Background()
    engine := createEngine()
    
    game, _ := engine.CreateGame(ctx, "alice", "bob")
    
    // Fill column 0 completely (6 discs)
    players := []string{"alice", "bob"}
    for i := 0; i < 6; i++ {
        engine.MakeMove(ctx, game.ID, players[i%2], 0)
    }
    
    // Try to add 7th disc to column 0
    err := engine.ValidateMove(ctx, game.ID, "alice", 0)
    if err != nil {
        fmt.Printf("✗ Invalid move: %v\n", err)
        // Output: ✗ Invalid move: column 0 is full
    }
}
```

### Wrong Turn

```go
func demonstrateWrongTurn() {
    ctx := context.Background()
    engine := createEngine()
    
    game, _ := engine.CreateGame(ctx, "alice", "bob")
    
    // Bob tries to move when it's Alice's turn
    err := engine.ValidateMove(ctx, game.ID, "bob", 3)
    if err != nil {
        fmt.Printf("✗ Invalid move: %v\n", err)
        // Output: ✗ Invalid move: it's not bob's turn
    }
}
```


## Game Session Management

### Checking Game Status

```go
func checkGameStatus(engine game.Engine, gameID string) {
    ctx := context.Background()
    
    game, err := engine.GetGame(ctx, gameID)
    if err != nil {
        log.Printf("Failed to get game: %v", err)
        return
    }
    
    fmt.Printf("Game ID: %s\n", game.ID)
    fmt.Printf("Status: %s\n", game.Status)
    fmt.Printf("Current Turn: %s\n", game.CurrentTurn)
    fmt.Printf("Is Active: %v\n", game.IsActive())
    
    if game.Winner != nil {
        fmt.Printf("Winner: %s\n", *game.Winner)
    }
}
```

### Game End Detection with Engine

```go
func demonstrateGameEndDetection() {
    ctx := context.Background()
    engine := createEngine()
    
    game, _ := engine.CreateGame(ctx, "alice", "bob")
    
    // Play moves that lead to a win
    moves := []struct {
        player string
        column int
    }{
        {"alice", 0}, {"bob", 1},
        {"alice", 0}, {"bob", 1},
        {"alice", 0}, {"bob", 1},
        {"alice", 0}, // Alice wins vertically!
    }
    
    for _, move := range moves {
        result, err := engine.MakeMove(ctx, game.ID, move.player, move.column)
        if err != nil {
            log.Printf("Move failed: %v", err)
            return
        }
        
        if result.GameEnded {
            if result.IsDraw {
                fmt.Println("Game ended in a DRAW!")
            } else if result.Winner != nil {
                fmt.Printf("Game ended! Winner: %s\n", *result.Winner)
            }
            break
        }
    }
}
```


## Interactive Game Loop

### Console-Based Game

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strconv"
    "strings"
    
    "connect4-multiplayer/internal/game"
    "connect4-multiplayer/pkg/models"
)

func runInteractiveGame() {
    ctx := context.Background()
    engine := createEngine()
    
    reader := bufio.NewReader(os.Stdin)
    
    fmt.Println("=== Connect 4 ===")
    fmt.Print("Player 1 name: ")
    player1, _ := reader.ReadString('\n')
    player1 = strings.TrimSpace(player1)
    
    fmt.Print("Player 2 name: ")
    player2, _ := reader.ReadString('\n')
    player2 = strings.TrimSpace(player2)
    
    gameSession, err := engine.CreateGame(ctx, player1, player2)
    if err != nil {
        fmt.Printf("Failed to create game: %v\n", err)
        return
    }
    
    fmt.Printf("\nGame started! %s (R) vs %s (Y)\n", player1, player2)
    printBoard(gameSession.Board)
    
    for {
        // Get current player
        currentPlayer := gameSession.GetCurrentPlayer()
        fmt.Printf("\n%s's turn (enter column 0-6): ", currentPlayer)
        
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)
        
        column, err := strconv.Atoi(input)
        if err != nil {
            fmt.Println("Invalid input. Please enter a number 0-6.")
            continue
        }
        
        result, err := engine.MakeMove(ctx, gameSession.ID, currentPlayer, column)
        if err != nil {
            fmt.Printf("Invalid move: %v\n", err)
            continue
        }
        
        gameSession = result.GameSession
        printBoard(gameSession.Board)
        
        if result.GameEnded {
            if result.IsDraw {
                fmt.Println("\n🤝 It's a DRAW!")
            } else {
                winner := gameSession.GetCurrentPlayer()
                if result.Winner != nil {
                    if *result.Winner == models.PlayerColorRed {
                        winner = player1
                    } else {
                        winner = player2
                    }
                }
                fmt.Printf("\n🎉 %s WINS!\n", winner)
            }
            break
        }
    }
}
```


## Testing Examples

### Unit Test for Win Detection

```go
package game_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "connect4-multiplayer/pkg/models"
)

func TestHorizontalWin(t *testing.T) {
    board := models.NewBoard()
    
    // Place 4 red discs horizontally
    for col := 0; col < 4; col++ {
        err := board.MakeMove(col, models.PlayerColorRed)
        assert.NoError(t, err)
    }
    
    winner := board.CheckWin()
    assert.NotNil(t, winner)
    assert.Equal(t, models.PlayerColorRed, *winner)
}

func TestVerticalWin(t *testing.T) {
    board := models.NewBoard()
    
    // Place 4 yellow discs vertically
    for i := 0; i < 4; i++ {
        err := board.MakeMove(3, models.PlayerColorYellow)
        assert.NoError(t, err)
    }
    
    winner := board.CheckWin()
    assert.NotNil(t, winner)
    assert.Equal(t, models.PlayerColorYellow, *winner)
}

func TestNoWinnerYet(t *testing.T) {
    board := models.NewBoard()
    
    // Place only 3 discs
    for col := 0; col < 3; col++ {
        board.MakeMove(col, models.PlayerColorRed)
    }
    
    winner := board.CheckWin()
    assert.Nil(t, winner)
}

func TestInvalidMove(t *testing.T) {
    board := models.NewBoard()
    
    // Fill a column
    for i := 0; i < 6; i++ {
        board.MakeMove(0, models.PlayerColorRed)
    }
    
    // Try to add to full column
    err := board.MakeMove(0, models.PlayerColorYellow)
    assert.Error(t, err)
    assert.Equal(t, models.ErrInvalidMove, err)
}
```


### Property-Based Test Example

```go
//go:build property
// +build property

package game_test

import (
    "testing"
    
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
    
    "connect4-multiplayer/pkg/models"
)

func TestWinDetectionProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Property: Any horizontal 4-in-a-row should be detected
    properties.Property("horizontal wins detected", prop.ForAll(
        func(row, startCol int, player models.PlayerColor) bool {
            if row < 0 || row >= 6 || startCol < 0 || startCol > 3 {
                return true
            }
            
            board := models.NewBoard()
            for i := 0; i < 4; i++ {
                board.Grid[row][startCol+i] = player
            }
            
            winner := board.CheckWin()
            return winner != nil && *winner == player
        },
        gen.IntRange(0, 5),
        gen.IntRange(0, 3),
        gen.OneConstOf(models.PlayerColorRed, models.PlayerColorYellow),
    ))
    
    // Property: Empty board should have no winner
    properties.Property("empty board has no winner", prop.ForAll(
        func() bool {
            board := models.NewBoard()
            return board.CheckWin() == nil
        },
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

## Helper Functions

```go
// createEngine creates a game engine with mock repositories
func createEngine() game.Engine {
    gameRepo := NewMockGameSessionRepository()
    moveRepo := NewMockMoveRepository()
    return game.NewEngine(gameRepo, moveRepo)
}

// printBoard prints the board state to console
func printBoard(board models.Board) {
    fmt.Println("\n  0 1 2 3 4 5 6")
    fmt.Println("  ─────────────")
    for row := 5; row >= 0; row-- {
        fmt.Printf("%d│", row)
        for col := 0; col < 7; col++ {
            switch board.Grid[row][col] {
            case models.PlayerColorRed:
                fmt.Print("R ")
            case models.PlayerColorYellow:
                fmt.Print("Y ")
            default:
                fmt.Print(". ")
            }
        }
        fmt.Println("│")
    }
    fmt.Println("  ─────────────")
}
```


## API Integration Example

### REST API Handler

```go
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "connect4-multiplayer/internal/game"
)

type GameHandler struct {
    engine game.Engine
}

func NewGameHandler(engine game.Engine) *GameHandler {
    return &GameHandler{engine: engine}
}

// CreateGame handles POST /api/v1/games
func (h *GameHandler) CreateGame(c *gin.Context) {
    var req struct {
        Player1 string `json:"player1" binding:"required"`
        Player2 string `json:"player2" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    game, err := h.engine.CreateGame(c.Request.Context(), req.Player1, req.Player2)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, game)
}

// MakeMove handles POST /api/v1/games/:id/moves
func (h *GameHandler) MakeMove(c *gin.Context) {
    gameID := c.Param("id")
    
    var req struct {
        Player string `json:"player" binding:"required"`
        Column int    `json:"column" binding:"min=0,max=6"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    result, err := h.engine.MakeMove(c.Request.Context(), gameID, req.Player, req.Column)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, result)
}

// GetGame handles GET /api/v1/games/:id
func (h *GameHandler) GetGame(c *gin.Context) {
    gameID := c.Param("id")
    
    game, err := h.engine.GetGame(c.Request.Context(), gameID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, game)
}
```


## WebSocket Integration Example

```go
package websocket

import (
    "context"
    "encoding/json"
    
    "github.com/gorilla/websocket"
    "connect4-multiplayer/internal/game"
)

type GameWebSocket struct {
    engine game.Engine
    conn   *websocket.Conn
    gameID string
}

type WSMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

type MovePayload struct {
    Player string `json:"player"`
    Column int    `json:"column"`
}

func (ws *GameWebSocket) HandleMessage(msg WSMessage) error {
    switch msg.Type {
    case "make_move":
        var payload MovePayload
        if err := json.Unmarshal(msg.Payload, &payload); err != nil {
            return err
        }
        
        result, err := ws.engine.MakeMove(
            context.Background(),
            ws.gameID,
            payload.Player,
            payload.Column,
        )
        if err != nil {
            return ws.sendError(err)
        }
        
        // Broadcast move to all connected clients
        return ws.broadcast("move_made", result)
        
    case "get_state":
        game, err := ws.engine.GetGame(context.Background(), ws.gameID)
        if err != nil {
            return ws.sendError(err)
        }
        return ws.send("game_state", game)
    }
    
    return nil
}

func (ws *GameWebSocket) send(msgType string, payload interface{}) error {
    data, _ := json.Marshal(payload)
    return ws.conn.WriteJSON(WSMessage{
        Type:    msgType,
        Payload: data,
    })
}

func (ws *GameWebSocket) sendError(err error) error {
    return ws.send("error", map[string]string{"message": err.Error()})
}

func (ws *GameWebSocket) broadcast(msgType string, payload interface{}) error {
    // In real implementation, broadcast to all game participants
    return ws.send(msgType, payload)
}
```

## Running the Examples

```bash
# Run unit tests
go test ./internal/game -v

# Run property-based tests
go test -tags=property ./internal/game -v

# Run the server
go run cmd/server/main.go

# Test API endpoints
curl -X POST http://localhost:8080/api/v1/games \
  -H "Content-Type: application/json" \
  -d '{"player1": "alice", "player2": "bob"}'

curl -X POST http://localhost:8080/api/v1/games/{gameID}/moves \
  -H "Content-Type: application/json" \
  -d '{"player": "alice", "column": 3}'
```

## Summary

This implementation guide demonstrates:

1. **Game Creation** - How to initialize games between players
2. **Move Processing** - Validating and executing moves
3. **Win Detection** - All four winning directions
4. **Draw Detection** - Full board without winner
5. **Error Handling** - Invalid moves and edge cases
6. **Testing** - Unit and property-based tests
7. **API Integration** - REST and WebSocket handlers

The game engine follows clean architecture principles with clear separation between game logic, data access, and presentation layers.
