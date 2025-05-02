package middleware

import (
	"starter-app/pkg/env"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// LimiterMiddleware creates a rate limiter middleware to prevent brute force attacks
func LimiterMiddleware() fiber.Handler {
	// Get rate limit config from environment or use defaults
	maxRequests := env.GetInt("RATE_LIMIT_MAX", 5)

	// Parse duration string like "1m" or use default 1 minute
	windowStr := env.Get("RATE_LIMIT_WINDOW", "1m")
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		window = time.Minute // Default to 1 minute if invalid
	}

	return limiter.New(limiter.Config{
		Max:               maxRequests,
		Expiration:        window,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c fiber.Ctx) string {
			// Use the IP address as the key
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "Rate limit exceeded, please try again later",
			})
		},
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Add security headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer-when-downgrade")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self';")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Continue to the next handler
		return c.Next()
	}
}

// ContentTypeMiddleware enforces correct content types
func ContentTypeMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Skip for non-API requests like static files
		if !strings.HasPrefix(c.Path(), "/api") && !strings.HasPrefix(c.Path(), "/auth") {
			return c.Next()
		}

		// Set content type for JSON responses
		c.Set("Content-Type", "application/json")

		// Validate content type for POST, PUT, PATCH requests
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"message": "Content-Type must be application/json",
				})
			}
		}

		return c.Next()
	}
}
