package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// TokenRepo handles refresh token persistence.
type TokenRepo struct {
	pool *pgxpool.Pool
}

// NewTokenRepo creates a new TokenRepo.
func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

// Create inserts a new refresh token.
func (r *TokenRepo) Create(ctx context.Context, t *models.RefreshToken) (*models.RefreshToken, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		t.UserID, t.TokenHash, t.ExpiresAt,
	).Scan(&t.ID, &t.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return t, nil
}

// FindByHash looks up a refresh token by its hash. Returns nil, nil if not found.
func (r *TokenRepo) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	t := &models.RefreshToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token by hash: %w", err)
	}
	return t, nil
}

// Delete removes a refresh token by its ID.
func (r *TokenRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteByUserID removes all refresh tokens for a given user.
func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete refresh tokens by user: %w", err)
	}
	return nil
}
