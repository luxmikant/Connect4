package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationConfig holds validation middleware configuration
type ValidationConfig struct {
	Validator *validator.Validate
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Validator: validator.New(),
	}
}

// Validation returns a middleware that sets up request validation
func Validation(config *ValidationConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultValidationConfig()
	}

	return func(c *gin.Context) {
		// Store validator in context for handlers to use
		c.Set("validator", config.Validator)
		c.Next()
	}
}

// ValidateJSON validates JSON request body against a struct
func ValidateJSON(c *gin.Context, obj interface{}) error {
	// Bind JSON to struct
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
		return err
	}

	// Get validator from context
	validatorInterface, exists := c.Get("validator")
	if !exists {
		// Fallback to default validator
		validatorInterface = validator.New()
	}
	
	v := validatorInterface.(*validator.Validate)

	// Validate struct
	if err := v.Struct(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return err
	}

	return nil
}