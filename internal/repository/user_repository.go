package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"starter-app/internal/model"
	"starter-app/pkg/cache"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db    *pgxpool.Pool
	cache cache.Store
	ttl   time.Duration
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool, cacheStore cache.Store, ttl time.Duration) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cacheStore,
		ttl:   ttl,
	}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	// Try to get from cache first
	if r.cache != nil {
		if data, found := r.cache.Get(cacheKey); found {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// Cache miss, query the database
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No user found
		}
		return nil, err
	}

	// Store in cache if found
	if r.cache != nil {
		if data, err := json.Marshal(user); err == nil {
			r.cache.Set(cacheKey, data, r.ttl)
		}
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)

	// Try from cache first
	if r.cache != nil {
		if data, found := r.cache.Get(cacheKey); found {
			var user model.User
			if err := json.Unmarshal(data, &user); err == nil {
				return &user, nil
			}
		}
	}

	// Cache miss or error, query the database
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No user found
		}
		return nil, err
	}

	// Store in cache if found
	if r.cache != nil {
		if data, err := json.Marshal(user); err == nil {
			r.cache.Set(cacheKey, data, r.ttl)
			// Also cache by ID
			r.cache.Set(fmt.Sprintf("user:%s", user.ID), data, r.ttl)
		}
	}

	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	// Invalidate cache
	if r.cache != nil {
		r.cache.Delete(fmt.Sprintf("user:%s", user.ID))
		r.cache.Delete(fmt.Sprintf("user:email:%s", user.Email))
	}

	return err
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password = $3, role = $4, updated_at = $5
		WHERE id = $6
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	// Invalidate cache
	if r.cache != nil {
		r.cache.Delete(fmt.Sprintf("user:%s", user.ID))
		r.cache.Delete(fmt.Sprintf("user:email:%s", user.Email))
	}

	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	// Get user first to get email for cache invalidation
	user, _ := r.GetByID(ctx, id)

	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)

	// Invalidate cache
	if user != nil && r.cache != nil {
		r.cache.Delete(fmt.Sprintf("user:%s", id))
		r.cache.Delete(fmt.Sprintf("user:email:%s", user.Email))
	}

	return err
}

// GetAll retrieves all users
func (r *UserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
