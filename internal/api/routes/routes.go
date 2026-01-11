package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"connect4-multiplayer/internal/api/handlers"
	"connect4-multiplayer/internal/api/middleware"
	"connect4-multiplayer/internal/auth"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/websocket"
)

// SetupRoutes configures all API routes and middleware
func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	gameHandler *handlers.GameHandler,
	leaderboardHandler *handlers.LeaderboardHandler,
	authHandler *handlers.AuthHandler,
	wsHandler *websocket.WebSocketHandler,
	supabaseAuth *auth.SupabaseAuth,
) {
	// Setup middleware
	setupMiddleware(router, cfg)

	// Setup API routes
	setupAPIRoutes(router, gameHandler, leaderboardHandler, authHandler, supabaseAuth)

	// Setup WebSocket routes
	setupWebSocketRoutes(router, wsHandler)

	// Setup documentation
	setupDocumentation(router)
}

// setupMiddleware configures all middleware
func setupMiddleware(router *gin.Engine, cfg *config.Config) {
	// CORS middleware
	corsConfig := &middleware.CORSConfig{
		AllowOrigins: cfg.Server.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Accept",
			"Cache-Control",
			"X-Requested-With",
		},
	}
	router.Use(middleware.CORS(corsConfig))

	// Logging middleware
	loggingConfig := middleware.DefaultLoggingConfig()
	router.Use(middleware.Logging(loggingConfig))

	// Recovery middleware
	recoveryConfig := middleware.DefaultRecoveryConfig()
	router.Use(middleware.Recovery(recoveryConfig))

	// Validation middleware
	validationConfig := middleware.DefaultValidationConfig()
	router.Use(middleware.Validation(validationConfig))
}

// setupAPIRoutes configures all API routes
func setupAPIRoutes(
	router *gin.Engine,
	gameHandler *handlers.GameHandler,
	leaderboardHandler *handlers.LeaderboardHandler,
	authHandler *handlers.AuthHandler,
	supabaseAuth *auth.SupabaseAuth,
) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "connect4-multiplayer",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication endpoints
		authGroup := v1.Group("/auth")
		authGroup.Use(middleware.SupabaseAuthMiddleware(supabaseAuth))
		{
			authGroup.GET("/me", authHandler.GetMe)
			authGroup.PUT("/profile", authHandler.UpdateProfile)
			authGroup.POST("/player", authHandler.GetOrCreatePlayer)
		}

		// Game management endpoints
		games := v1.Group("/games")
		{
			games.POST("", gameHandler.CreateGame)
			games.GET("/:id", gameHandler.GetGameState)
			games.POST("/:id/moves", gameHandler.MakeMove)
		}

		// Leaderboard endpoints
		v1.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		// Player statistics endpoints
		players := v1.Group("/players")
		{
			players.GET("/:id/stats", leaderboardHandler.GetPlayerStats)
		}
	}
}

// setupDocumentation configures Swagger documentation
func setupDocumentation(router *gin.Engine) {
	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupWebSocketRoutes configures WebSocket routes
func setupWebSocketRoutes(router *gin.Engine, wsHandler *websocket.WebSocketHandler) {
	// WebSocket endpoint for real-time game communication
	router.GET("/ws", wsHandler.HandleWebSocket)
}
