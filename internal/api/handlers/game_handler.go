package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"connect4-multiplayer/internal/game"
)

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	gameService game.GameService
	validator   *validator.Validate
}

// NewGameHandler creates a new GameHandler instance
func NewGameHandler(gameService game.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		validator:   validator.New(),
	}
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	Player1 string `json:"player1" validate:"required,min=3,max=20"`
	Player2 string `json:"player2" validate:"required,min=3,max=20"`
}

// MakeMoveRequest represents the request to make a move
type MakeMoveRequest struct {
	Column int    `json:"column" validate:"required,min=0,max=6"`
	Player string `json:"player" validate:"required,min=3,max=20"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// CreateGame creates a new Connect 4 game session
// @Summary Create new game
// @Description Create a new Connect 4 game session between two players
// @Tags games
// @Accept json
// @Produce json
// @Param request body CreateGameRequest true "Game creation request"
// @Success 201 {object} models.GameSession
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games [post]
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Check if players have different usernames
	if req.Player1 == req.Player2 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Players must have different usernames",
		})
		return
	}

	// Create game session
	session, err := h.gameService.CreateSession(c.Request.Context(), req.Player1, req.Player2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to create game session",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetGameState retrieves the current state of a game
// @Summary Get game state
// @Description Retrieve the current state of a Connect 4 game session
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Success 200 {object} models.GameSession
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id} [get]
func (h *GameHandler) GetGameState(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Game ID is required",
		})
		return
	}

	session, err := h.gameService.GetSession(c.Request.Context(), gameID)
	if err != nil {
		if err.Error() == "game session not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Game not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve game state",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// MakeMove makes a move in a Connect 4 game
// @Summary Make a move
// @Description Make a move in a Connect 4 game session
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Param request body MakeMoveRequest true "Move request"
// @Success 200 {object} models.GameSession
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /games/{id}/moves [post]
func (h *GameHandler) MakeMove(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Game ID is required",
		})
		return
	}

	var req MakeMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Get current game session
	session, err := h.gameService.GetSession(c.Request.Context(), gameID)
	if err != nil {
		if err.Error() == "game session not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Game not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve game state",
			Details: err.Error(),
		})
		return
	}

	// Check if game is active
	if !session.IsActive() {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: "Game is not active",
		})
		return
	}

	// Check if it's the player's turn
	currentPlayer := session.GetCurrentPlayer()
	if currentPlayer != req.Player {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: "Not your turn",
		})
		return
	}

	// Validate move
	if !session.Board.IsValidMove(req.Column) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid move: column is full or out of bounds",
		})
		return
	}

	// Make the move
	playerColor := session.GetPlayerColor(req.Player)
	if err := session.Board.MakeMove(req.Column, playerColor); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Failed to make move",
			Details: err.Error(),
		})
		return
	}

	// Check for win or draw
	winner := session.Board.CheckWin()
	if winner != nil {
		// Game won
		if err := h.gameService.CompleteGame(c.Request.Context(), gameID, winner); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to complete game",
				Details: err.Error(),
			})
			return
		}
	} else if session.Board.IsFull() {
		// Game is a draw
		if err := h.gameService.CompleteGame(c.Request.Context(), gameID, nil); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to complete game",
				Details: err.Error(),
			})
			return
		}
	} else {
		// Switch turn and update session
		if err := h.gameService.SwitchTurn(c.Request.Context(), gameID); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to switch turn",
				Details: err.Error(),
			})
			return
		}
	}

	// Get updated session
	updatedSession, err := h.gameService.GetSession(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve updated game state",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedSession)
}