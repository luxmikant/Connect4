package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"connect4-multiplayer/internal/database/repositories"
)

// LeaderboardHandler handles leaderboard-related HTTP requests
type LeaderboardHandler struct {
	statsRepo repositories.PlayerStatsRepository
}

// NewLeaderboardHandler creates a new LeaderboardHandler instance
func NewLeaderboardHandler(statsRepo repositories.PlayerStatsRepository) *LeaderboardHandler {
	return &LeaderboardHandler{
		statsRepo: statsRepo,
	}
}

// GetLeaderboard retrieves the top players leaderboard
// @Summary Get leaderboard
// @Description Retrieve the top players ranked by wins
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param limit query int false "Number of players to return (default: 10, max: 100)"
// @Success 200 {array} models.PlayerStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /leaderboard [get]
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	
	// Cap the limit to prevent excessive queries
	if limit > 100 {
		limit = 100
	}

	// Get leaderboard data
	leaderboard, err := h.statsRepo.GetLeaderboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve leaderboard",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetPlayerStats retrieves statistics for a specific player
// @Summary Get player statistics
// @Description Retrieve detailed statistics for a specific player
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player username"
// @Success 200 {object} models.PlayerStats
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /players/{id}/stats [get]
func (h *LeaderboardHandler) GetPlayerStats(c *gin.Context) {
	username := c.Param("id")
	if username == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Player username is required",
		})
		return
	}

	// Get player statistics
	stats, err := h.statsRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		if err.Error() == "player stats not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Player not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve player statistics",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}