package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// NoteRepo handles note persistence.
type NoteRepo struct {
	pool *pgxpool.Pool
}

// NewNoteRepo creates a new NoteRepo.
func NewNoteRepo(pool *pgxpool.Pool) *NoteRepo {
	return &NoteRepo{pool: pool}
}

// Get retrieves the note for a user on a given date. Returns nil, nil if not found.
func (r *NoteRepo) Get(ctx context.Context, userID, date string) (*models.Note, error) {
	n := &models.Note{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, date, text FROM notes WHERE user_id = $1 AND date = $2`,
		userID, date,
	).Scan(&n.ID, &n.UserID, &n.Date, &n.Text)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get note: %w", err)
	}
	return n, nil
}

// Upsert creates or updates a note for a user on a given date.
func (r *NoteRepo) Upsert(ctx context.Context, userID, date, text string) (*models.Note, error) {
	n := &models.Note{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO notes (user_id, date, text)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, date) DO UPDATE SET text = EXCLUDED.text
		 RETURNING id, user_id, date, text`,
		userID, date, text,
	).Scan(&n.ID, &n.UserID, &n.Date, &n.Text)

	if err != nil {
		return nil, fmt.Errorf("upsert note: %w", err)
	}
	return n, nil
}
