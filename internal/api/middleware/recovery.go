package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryConfig holds recovery middleware configuration
type RecoveryConfig struct {
	Logger *slog.Logger
}

// DefaultRecoveryConfig returns default recovery configuration
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		Logger: slog.Default(),
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery(config *RecoveryConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRecoveryConfig()
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				config.Logger.Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", string(debug.Stack()),
				)

				// Return internal server error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
				
				c.Abort()
			}
		}()
		
		c.Next()
	}
}