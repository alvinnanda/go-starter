package repository

import (
	"context"
	"errors"
	"fmt"
	"starter-app/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository handles database operations for refresh tokens
type TokenRepository struct {
	db *pgxpool.Pool
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// Create stores a new refresh token
func (r *TokenRepository) Create(ctx context.Context, userID, token string, expiresAt time.Time) error {
	// First, try to delete any existing tokens with the same value to avoid conflicts
	deleteQuery := `DELETE FROM refresh_tokens WHERE token = $1`
	_, _ = r.db.Exec(ctx, deleteQuery, token)

	// Then insert the new token
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

// GetByToken retrieves a refresh token by its token string
func (r *TokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var refreshToken model.RefreshToken
	err := r.db.QueryRow(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No token found
		}
		return nil, err
	}

	return &refreshToken, nil
}

// Delete removes a refresh token
func (r *TokenRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.Exec(ctx, query, token)
	return err
}

// DeleteAllForUser removes all refresh tokens for a user
func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// DeleteExpired removes all expired refresh tokens
func (r *TokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)
	return err
}
