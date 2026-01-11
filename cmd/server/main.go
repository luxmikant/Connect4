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

	"connect4-multiplayer/internal/api/handlers"
	"connect4-multiplayer/internal/api/routes"
	"connect4-multiplayer/internal/config"
	"connect4-multiplayer/internal/database"
	"connect4-multiplayer/internal/game"
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

	// Initialize database
	_, repoManager, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	gameService := game.NewGameService(
		repoManager.GameSession,
		repoManager.PlayerStats,
		repoManager.Move,
		repoManager.GameEvent,
		game.DefaultServiceConfig(),
	)

	// Initialize WebSocket service
	wsService := websocket.NewService(gameService)
	
	// Start WebSocket service
	ctx := context.Background()
	if err := wsService.Start(ctx); err != nil {
		log.Fatalf("Failed to start WebSocket service: %v", err)
	}

	// Initialize handlers
	gameHandler := handlers.NewGameHandler(gameService)
	leaderboardHandler := handlers.NewLeaderboardHandler(repoManager.PlayerStats)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Setup routes and middleware
	routes.SetupRoutes(router, cfg, gameHandler, leaderboardHandler, wsService.GetWebSocketHandler())

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
	if err := wsService.Stop(); err != nil {
		log.Printf("Error stopping WebSocket service: %v", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}