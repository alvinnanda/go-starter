package middleware

import (
	"fmt"
	"starter-app/config"
	"starter-app/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// JWTAuthMiddleware creates a middleware to protect routes with JWT authentication
func JWTAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")

		// Check if the header is empty
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Authorization header is missing",
			})
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid authorization format, expected Bearer token",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := jwt.ValidateToken(tokenString, cfg.Auth.JWTSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid or expired token: " + err.Error(),
			})
		}

		// Set user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		// Continue to the next middleware/handler
		return c.Next()
	}
}

// RoleGuard creates a middleware to restrict routes based on user role
func RoleGuard(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get the user role from context
		roleInterface := c.Locals("role")
		if roleInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Access denied: role not found in context",
			})
		}

		// Convert interface to string
		role, ok := roleInterface.(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Access denied: invalid role format",
			})
		}

		fmt.Printf("User role: %s, Required roles: %v\n", role, roles)

		// Check if the user role is in the allowed roles
		allowed := false
		for _, r := range roles {
			if r == role {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Access denied: insufficient permissions",
			})
		}

		// Continue to the next middleware/handler
		return c.Next()
	}
}
