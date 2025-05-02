package main

import (
	"log"
	"starter-app/config"
	"starter-app/internal/middleware"
	"starter-app/internal/repository"
	"starter-app/pkg/cache"
	"starter-app/pkg/database"
	"starter-app/pkg/env"
	"starter-app/pkg/logger"
	"starter-app/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	logger.Info("Starting application...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config:", err)
		return
	}

	// Initialize database connection
	dbPool, err := database.NewPostgresPool(cfg)
	if err != nil {
		logger.Error("Failed to connect to database:", err)
		return
	}
	defer dbPool.Close()
	logger.Info("Connected to database")

	// Initialize cache
	cacheConfig := cache.DefaultConfig()
	cacheStore := cache.New(cacheConfig)
	logger.Info("Cache initialized with driver:", cacheConfig.Driver)

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool, cacheStore, cacheConfig.TTL)
	tokenRepo := repository.NewTokenRepository(dbPool)

	// Create a new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "Starter App",
		// Enhance security - prevent server info leakage
		ServerHeader: "Server",
		// Restrict body size to prevent DOS attacks
		BodyLimit: 10 * 1024 * 1024, // 10MB
		// Enable secure error handling
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Return a generic error message to prevent information leakage
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error("Error:", err)
			return c.Status(code).JSON(fiber.Map{
				"message": "An error occurred while processing your request",
			})
		},
	})

	// Add global middleware
	app.Use(recover.New())                          // Recover from panics
	app.Use(middleware.SecurityHeadersMiddleware()) // Add security headers
	app.Use(middleware.ContentTypeMiddleware())     // Validate content types

	// Configure CORS with secure defaults
	allowOrigins := []string{env.Get("ALLOWED_ORIGIN", "*")}
	allowCredentials := true
	if allowOrigins[0] == "*" {
		allowCredentials = false // Disable credentials if wildcard is used
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials,
		MaxAge:           86400, // 24 hours
	}))

	// Setup routes
	routes.Setup(app, cfg, userRepo, tokenRepo)

	// Add 404 handler - must be added after all other routes
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Route " + c.Path() + " not found",
		})
	})

	// Start server
	logger.Info("Server is starting on port:", cfg.Server.Port)
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
