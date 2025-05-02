package utils

import (
	"errors"
	"starter-app/pkg/env"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Errors for password validation
var (
	ErrPasswordTooShort    = errors.New("password is too short (minimum 8 characters)")
	ErrPasswordNoUpper     = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower     = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber    = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial   = errors.New("password must contain at least one special character")
	ErrPasswordCommonWords = errors.New("password contains common words or patterns")
)

// CommonPasswords contains a list of frequently used passwords to check against
var CommonPasswords = map[string]bool{
	"password": true, "123456": true, "qwerty": true, "admin": true,
	"welcome": true, "password123": true, "abc123": true, "letmein": true,
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	// Check for common passwords and patterns
	lowerPass := strings.ToLower(password)
	for common := range CommonPasswords {
		if strings.Contains(lowerPass, common) {
			return ErrPasswordCommonWords
		}
	}

	return nil
}

// HashPassword creates a secure password hash with proper cost
func HashPassword(password string) (string, error) {
	// Get bcrypt cost from environment or use default (12)
	cost := env.GetInt("BCRYPT_COST", 12)

	// Ensure cost is within bcrypt's allowed range (4-31)
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	} else if cost > 31 {
		cost = 31
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPasswordHash verifies a password against a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
