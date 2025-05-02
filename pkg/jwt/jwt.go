package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims represents JWT claims with user information
type CustomClaims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	TokenID string `json:"jti"` // Token ID for revocation
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID, email, role, secret string, expiry time.Duration) (string, error) {
	// Generate a unique token ID for revocation capability
	tokenID := uuid.New().String()

	// Create claims with user information
	claims := CustomClaims{
		UserID:  userID,
		Email:   email,
		Role:    role,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "starter-app",
			Subject:   userID,
			ID:        tokenID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Use a more secure approach with SHA-256
	hashedSecret := hashSecret(secret)

	// Generate the signed token
	tokenString, err := token.SignedString([]byte(hashedSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*CustomClaims, error) {
	// Hash the secret
	hashedSecret := hashSecret(secret)

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(hashedSecret), nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	// Validate the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Get the claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

// hashSecret creates a more secure version of the secret by hashing it
func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
