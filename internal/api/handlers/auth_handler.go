package handlers

import (
	"net/http"

	"connect4-multiplayer/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related endpoints
type AuthHandler struct {
	supabaseAuth *auth.SupabaseAuth
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(supabaseAuth *auth.SupabaseAuth) *AuthHandler {
	return &AuthHandler{
		supabaseAuth: supabaseAuth,
	}
}

// GetMe returns the authenticated user's profile
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profile, err := h.supabaseAuth.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the authenticated user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Filter allowed fields
	allowedUpdates := make(map[string]interface{})
	if val, ok := updates["username"]; ok {
		allowedUpdates["username"] = val
	}
	if val, ok := updates["avatar_url"]; ok {
		allowedUpdates["avatar_url"] = val
	}

	if len(allowedUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	err := h.supabaseAuth.UpdateProfile(c.Request.Context(), userID, allowedUpdates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
