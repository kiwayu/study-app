package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// UserRepo handles user persistence.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// FindByProviderID looks up a user by their OAuth provider and provider-specific ID.
// Returns nil, nil if not found.
func (r *UserRepo) FindByProviderID(ctx context.Context, provider, providerID string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE provider = $1 AND provider_id = $2`,
		provider, providerID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by provider: %w", err)
	}
	return u, nil
}

// Create inserts a new user and returns it with the generated ID and timestamps.
func (r *UserRepo) Create(ctx context.Context, u *models.User) (*models.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, avatar_url, provider, provider_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		u.Email, u.Name, u.AvatarURL, u.Provider, u.ProviderID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetByID retrieves a user by their UUID. Returns nil, nil if not found.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, avatar_url, provider, provider_id, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.ProviderID, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}
