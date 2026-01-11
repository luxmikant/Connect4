package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingConfig holds logging middleware configuration
type LoggingConfig struct {
	Logger     *slog.Logger
	SkipPaths  []string
	TimeFormat string
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Logger:     slog.Default(),
		SkipPaths:  []string{"/health", "/metrics"},
		TimeFormat: time.RFC3339,
	}
}

// Logging returns a middleware that logs HTTP requests
func Logging(config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		start := time.Now()
		
		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		
		// Get client IP
		clientIP := c.ClientIP()
		
		// Get request method and path
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		// Get user agent
		userAgent := c.Request.UserAgent()
		
		// Get request size
		requestSize := c.Request.ContentLength
		
		// Get response size
		responseSize := c.Writer.Size()

		// Log the request
		config.Logger.Info("HTTP request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"client_ip", clientIP,
			"user_agent", userAgent,
			"request_size", requestSize,
			"response_size", responseSize,
		)

		// Log errors if status code indicates an error
		if statusCode >= 400 {
			errors := c.Errors.String()
			if errors != "" {
				config.Logger.Error("HTTP request error",
					"method", method,
					"path", path,
					"status", statusCode,
					"errors", errors,
				)
			}
		}
	}
}