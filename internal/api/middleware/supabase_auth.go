package middleware

import (
	"net/http"
	"strings"

	"connect4-multiplayer/internal/auth"

	"github.com/gin-gonic/gin"
)

// SupabaseAuthMiddleware validates Supabase JWT tokens
func SupabaseAuthMiddleware(supabaseAuth *auth.SupabaseAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Verify token with Supabase
		user, err := supabaseAuth.VerifyToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user info in context for use in handlers
		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		c.Set("user", user)

		c.Next()
	}
}

// OptionalAuthMiddleware allows both authenticated and guest users
func OptionalAuthMiddleware(supabaseAuth *auth.SupabaseAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth provided, continue as guest
			c.Set("isGuest", true)
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue as guest
			c.Set("isGuest", true)
			c.Next()
			return
		}

		token := parts[1]
		user, err := supabaseAuth.VerifyToken(c.Request.Context(), token)
		if err != nil {
			// Invalid token, continue as guest
			c.Set("isGuest", true)
			c.Next()
			return
		}

		// Valid token, set user info
		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		c.Set("user", user)
		c.Set("isGuest", false)

		c.Next()
	}
}
