package routes

import (
	"starter-app/config"
	"starter-app/internal/handler"
	"starter-app/internal/middleware"
	"starter-app/internal/repository"

	"github.com/gofiber/fiber/v3"
)

// Setup sets up all the routes for the application
func Setup(app *fiber.App, cfg *config.Config, userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository) {
	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, userRepo, tokenRepo)
	userHandler := handler.NewUserHandler(userRepo)

	// Health check route
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Auth routes - add rate limiter to prevent brute force
	auth := app.Group("/auth")

	// Limited routes for login and register to prevent brute force attacks
	limitedAuth := auth.Group("/", middleware.LimiterMiddleware())
	limitedAuth.Post("/register", authHandler.Register)
	limitedAuth.Post("/login", authHandler.Login)
	limitedAuth.Post("/refresh", authHandler.RefreshToken)

	// Logout doesn't need limiting
	auth.Post("/logout", authHandler.Logout)

	// Protected routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Public API routes
	v1.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to the API",
		})
	})

	// User routes - apply JWT middleware to all routes
	users := v1.Group("/users", middleware.JWTAuthMiddleware(cfg))

	// Routes accessible to all authenticated users
	users.Get("/me", authHandler.Me)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)

	// Admin-only routes with explicit role guard
	adminRoutes := users.Group("/", middleware.RoleGuard("admin"))
	adminRoutes.Get("/", userHandler.GetUsers)
	adminRoutes.Post("/", userHandler.CreateUser)
	adminRoutes.Delete("/:id", userHandler.DeleteUser)
}
