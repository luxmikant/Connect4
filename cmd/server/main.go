package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"connect4-multiplayer/internal/analytics"
	"connect4-multiplayer/internal/api/handlers"
	"connect4-multiplayer/internal/api/routes"
	"connect4-multiplayer/internal/auth"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
	"connect4-multiplayer/internal/game"
	"connect4-multiplayer/internal/matchmaking"
	"connect4-multiplayer/internal/websocket"
)

// @title Connect 4 Multiplayer API
// @version 1.0
// @description A real-time multiplayer Connect 4 game system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database with resilient connection (won't crash if DB is unavailable)
	resilientDB := database.NewResilientDB(cfg.Database)
	defer resilientDB.Close()

	// Run migrations if DB is already connected
	if resilientDB.IsConnected() {
		log.Println("Running database migrations...")
		migrator := database.NewMigrator(resilientDB.DB())
		if err := migrator.Up(); err != nil {
			log.Printf("Warning: Migration failed: %v", err)
			log.Println("Continuing with server startup...")
		} else {
			log.Println("Migrations completed successfully")
		}
	} else {
		log.Println("⚠️  Skipping migrations — database not yet available (will run after reconnection)")
	}

	// Initialize Kafka analytics producer (Requirement 9)
	// Skip Kafka if credentials are not configured
	var analyticsProducer *analytics.Producer
	if cfg.Kafka.APIKey != "" && cfg.Kafka.APISecret != "" && cfg.Kafka.BootstrapServers != "" {
		analyticsProducer = analytics.NewProducer(cfg.Kafka)
		log.Printf("Analytics producer initialized for Kafka topic: %s", cfg.Kafka.Topic)
	} else {
		log.Println("Kafka credentials not configured, using noop analytics producer")
		analyticsProducer = analytics.NewNoopProducer()
	}

	// Initialize services with analytics producer
	serviceConfig := game.DefaultServiceConfig()
	serviceConfig.AnalyticsProducer = analyticsProducer

	// Build services — use repo manager if DB is connected, otherwise defer
	var gameService game.GameService
	var matchmakingService matchmaking.MatchmakingService
	var wsService *websocket.Service

	initServices := func() {
		repoManager := resilientDB.RepoManager()
		if repoManager == nil {
			log.Println("⚠️  Cannot initialize game services — database not connected")
			return
		}

		gameService = game.NewGameService(
			repoManager.GameSession,
			repoManager.PlayerStats,
			repoManager.Move,
			repoManager.GameEvent,
			serviceConfig,
		)

		matchmakingService = matchmaking.NewMatchmakingService(
			gameService,
			matchmaking.DefaultServiceConfig(),
		)

		wsService = websocket.NewService(gameService, matchmakingService)

		ctx := context.Background()
		if err := wsService.Start(ctx); err != nil {
			log.Printf("Warning: Failed to start WebSocket service: %v", err)
		} else {
			log.Println("✅ Game services initialized successfully")
		}
	}

	initServices()

	// If DB wasn't ready at startup, watch for reconnection and init services
	if !resilientDB.IsConnected() {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				if resilientDB.IsConnected() && gameService == nil {
					log.Println("Database reconnected — initializing game services...")
					initServices()
					if gameService != nil {
						break
					}
				}
			}
		}()
	}

	// Initialize Supabase Auth
	supabaseAuth := auth.NewSupabaseAuth(cfg.Supabase.URL, cfg.Supabase.ServiceKey)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Build handlers — may be nil if DB isn't ready yet
	var gameHandler *handlers.GameHandler
	var leaderboardHandler *handlers.LeaderboardHandler
	var authHandler *handlers.AuthHandler

	if resilientDB.IsConnected() {
		repoManager := resilientDB.RepoManager()
		gameHandler = handlers.NewGameHandler(gameService)
		leaderboardHandler = handlers.NewLeaderboardHandler(repoManager.PlayerStats)
		authHandler = handlers.NewAuthHandler(supabaseAuth, repoManager.Player)
	}

	// Setup routes — pass whatever we have (handlers may be nil initially)
	if gameHandler != nil && leaderboardHandler != nil && authHandler != nil && wsService != nil {
		routes.SetupRoutes(router, cfg, gameHandler, leaderboardHandler, authHandler, wsService.GetWebSocketHandler(), supabaseAuth)
	} else {
		// Minimal routes when DB is unavailable
		setupMinimalRoutes(router, resilientDB)
		log.Println("⚠️  Server started with minimal routes — game endpoints unavailable until DB connects")

		// Watch for DB reconnection and set up full routes
		go func() {
			for {
				time.Sleep(5 * time.Second)
				if resilientDB.IsConnected() && gameService != nil {
					repoManager := resilientDB.RepoManager()
					gh := handlers.NewGameHandler(gameService)
					lh := handlers.NewLeaderboardHandler(repoManager.PlayerStats)
					ah := handlers.NewAuthHandler(supabaseAuth, repoManager.Player)
					routes.SetupRoutes(router, cfg, gh, lh, ah, wsService.GetWebSocketHandler(), supabaseAuth)
					log.Println("✅ Full routes registered after DB reconnection")
					break
				}
			}
		}()
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop WebSocket service
	if wsService != nil {
		if err := wsService.Stop(); err != nil {
			log.Printf("Error stopping WebSocket service: %v", err)
		}
	}

	// Close analytics producer
	if err := analyticsProducer.Close(); err != nil {
		log.Printf("Error closing analytics producer: %v", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupMinimalRoutes registers only health/status routes when the database is unavailable.
// This keeps the server responsive to Render's health checks so it doesn't get killed.
func setupMinimalRoutes(router *gin.Engine, rdb *database.ResilientDB) {
	// Setup middleware even in minimal mode
	corsConfig := gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.Use(corsConfig)
	router.Use(gin.Recovery())

	// Health check — always returns 200 so Render doesn't kill the service
	router.GET("/health", func(c *gin.Context) {
		dbStatus := "disconnected"
		if rdb.IsConnected() {
			dbStatus = "connected"
		}
		c.JSON(200, gin.H{
			"status":   "degraded",
			"service":  "connect4-multiplayer",
			"version":  "1.0.0",
			"database": dbStatus,
			"message":  "Server is running but database-dependent endpoints are unavailable",
		})
	})

	// Readiness check — returns 503 until DB is connected
	router.GET("/health/ready", func(c *gin.Context) {
		if err := rdb.HealthCheck(); err != nil {
			c.JSON(503, gin.H{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Catch-all for API routes when DB is down
	router.NoRoute(func(c *gin.Context) {
		c.JSON(503, gin.H{
			"error":   "service_degraded",
			"message": "Database is currently unavailable. The server is reconnecting — please try again shortly.",
		})
	})
}
