package handler

import (
	"fmt"
	"starter-app/config"
	"starter-app/internal/model"
	"starter-app/internal/repository"
	"starter-app/internal/utils"
	"starter-app/pkg/jwt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	config    *config.Config
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository) *AuthHandler {
	return &AuthHandler{
		config:    cfg,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c fiber.Ctx) error {
	// Parse request body
	var req model.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name, email and password are required",
		})
	}

	// Validate password strength
	if err := utils.ValidatePassword(req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error checking existing user",
		})
	}

	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "User with this email already exists",
		})
	}

	// Hash password with more secure method
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error hashing password",
		})
	}

	// Create new user
	user := &model.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // Default role
	}

	// Save user to database
	if err := h.userRepo.Create(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating user",
		})
	}

	// Return success response without tokens
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully. Please login to continue.",
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	// Parse request body
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email and password are required",
		})
	}

	// Find user
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error finding user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials",
		})
	}

	// Check password using secure method
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials",
		})
	}

	// Generate tokens
	accessToken, err := jwt.GenerateToken(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.TokenExpiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating token",
		})
	}

	refreshToken, err := jwt.GenerateToken(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.RefreshExpiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating refresh token",
		})
	}

	// Store refresh token in database
	refreshExpiry := time.Now().Add(h.config.Auth.RefreshExpiry)
	if err := h.tokenRepo.Create(c.Context(), user.ID, refreshToken, refreshExpiry); err != nil {
		fmt.Printf("Error storing refresh token: %v\n", err)
		// Still return the token to the user, but log the error
		// This way users can still log in even if token storage fails
	}

	// Return tokens
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token": model.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(h.config.Auth.TokenExpiry.Seconds()),
			TokenType:    "Bearer",
		},
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(c fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID := c.Locals("user_id").(string)

	// Find user
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error finding user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Return user info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// RefreshToken refreshes an access token using a refresh token
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	// Parse request body
	var req model.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate refresh token exists
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Refresh token is required",
		})
	}

	// Verify the refresh token is valid
	claims, err := jwt.ValidateToken(req.RefreshToken, h.config.Auth.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid refresh token: " + err.Error(),
		})
	}

	// Check if the token exists in the database
	storedToken, err := h.tokenRepo.GetByToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error verifying refresh token",
		})
	}

	if storedToken == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Refresh token not found or expired",
		})
	}

	// Check if token has expired
	if storedToken.ExpiresAt.Before(time.Now()) {
		// Delete the expired token
		_ = h.tokenRepo.Delete(c.Context(), req.RefreshToken)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Refresh token has expired",
		})
	}

	// Get the user
	user, err := h.userRepo.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error finding user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Generate new tokens
	accessToken, err := jwt.GenerateToken(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.TokenExpiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating access token",
		})
	}

	newRefreshToken, err := jwt.GenerateToken(user.ID, user.Email, user.Role, h.config.Auth.JWTSecret, h.config.Auth.RefreshExpiry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error generating refresh token",
		})
	}

	// Delete the old refresh token
	if err := h.tokenRepo.Delete(c.Context(), req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error invalidating old refresh token",
		})
	}

	// Store the new refresh token
	refreshExpiry := time.Now().Add(h.config.Auth.RefreshExpiry)
	if err := h.tokenRepo.Create(c.Context(), user.ID, newRefreshToken, refreshExpiry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error storing new refresh token",
		})
	}

	// Return new tokens
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Token refreshed successfully",
		"token": model.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int64(h.config.Auth.TokenExpiry.Seconds()),
			TokenType:    "Bearer",
		},
	})
}

// Logout invalidates a refresh token
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	// Parse request body
	var req model.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate refresh token exists
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Refresh token is required",
		})
	}

	// Delete the refresh token
	if err := h.tokenRepo.Delete(c.Context(), req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error logging out",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
