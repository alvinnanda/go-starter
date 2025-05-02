package handler

import (
	"starter-app/internal/model"
	"starter-app/internal/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo *repository.UserRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GetUsers gets all users (admin only)
func (h *UserHandler) GetUsers(c fiber.Ctx) error {
	// Get all users from database
	users, err := h.userRepo.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error retrieving users",
			"error":   err.Error(),
		})
	}

	// Create response without exposing sensitive info
	var response []fiber.Map
	for _, user := range users {
		response = append(response, fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"users": response,
		"count": len(users),
	})
}

// GetUser gets a single user by ID
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Get the requesting user's ID
	currentUserID := c.Locals("user_id").(string)
	currentUserRole := c.Locals("role").(string)

	// Check if the user is accessing their own profile or is an admin
	if id != currentUserID && currentUserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have permission to access this user's information",
		})
	}

	// Get user from database
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error retrieving user",
			"error":   err.Error(),
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Return user without sensitive info
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error hashing password",
		})
	}

	// Get the requesting user's role
	currentUserRole := c.Locals("role").(string)

	// Create new user
	user := &model.User{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	// Set default role if none provided
	if user.Role == "" {
		user.Role = "user"
	} else if currentUserRole != "admin" {
		// If the requester is not an admin, override any role they might have set
		user.Role = "user"
	}

	// Save user to database
	if err := h.userRepo.Create(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creating user",
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user": fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Get the requesting user's ID
	currentUserID := c.Locals("user_id").(string)
	currentUserRole := c.Locals("role").(string)

	// Check if the user is updating their own profile or is an admin
	if id != currentUserID && currentUserRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have permission to update this user",
		})
	}

	// Get existing user
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error retrieving user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Parse request body
	type UpdateRequest struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		Role            string `json:"role"`
	}

	var req UpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Only allow fields that were provided in the request
	if req.Name != "" {
		user.Name = req.Name
	}

	// Email change requires special handling
	if req.Email != "" && req.Email != user.Email {
		// Only admin or the user themselves can change email
		if id != currentUserID && currentUserRole != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "You don't have permission to change this user's email",
			})
		}

		// Check if the new email is already in use
		existingUser, err := h.userRepo.GetByEmail(c.Context(), req.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error checking email availability",
			})
		}

		if existingUser != nil && existingUser.ID != id {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Email is already in use",
			})
		}

		user.Email = req.Email
	}

	// Password change requires current password verification
	if req.NewPassword != "" {
		// If it's not an admin, require current password
		if currentUserRole != "admin" {
			if req.CurrentPassword == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Current password is required to change password",
				})
			}

			// Verify current password
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Current password is incorrect",
				})
			}
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error hashing password",
			})
		}

		user.Password = string(hashedPassword)
	}

	// Role change requires admin privileges
	if req.Role != "" && req.Role != user.Role {
		if currentUserRole != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Only administrators can change user roles",
			})
		}

		// Prevent removing the last admin
		if user.Role == "admin" && req.Role != "admin" {
			// Count how many admins are in the system
			users, err := h.userRepo.GetAll(c.Context())
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Error checking admin count",
				})
			}

			adminCount := 0
			for _, u := range users {
				if u.Role == "admin" {
					adminCount++
				}
			}

			if adminCount <= 1 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Cannot demote the last administrator",
				})
			}
		}

		user.Role = req.Role
	}

	// Update user in database
	if err := h.userRepo.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"user": fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

// DeleteUser deletes a user (admin only)
func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}

	// Get the user from the database
	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error retrieving user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	// Prevent deleting the last admin
	if user.Role == "admin" {
		// Count how many admins are in the system
		users, err := h.userRepo.GetAll(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error checking admin count",
			})
		}

		adminCount := 0
		for _, u := range users {
			if u.Role == "admin" {
				adminCount++
			}
		}

		if adminCount <= 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Cannot delete the last administrator",
			})
		}
	}

	// Delete user from database
	if err := h.userRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error deleting user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
