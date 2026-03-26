package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// SessionRepo handles session state persistence.
type SessionRepo struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a new SessionRepo.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

// Get retrieves the session state for a user. Returns nil, nil if not found.
func (r *SessionRepo) Get(ctx context.Context, userID string) (*models.SessionState, error) {
	s := &models.SessionState{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, segment_type, segment_index, pomodoro_count,
		        started_at, elapsed_seconds, total_elapsed, last_water_at, last_stretch_at
		 FROM sessions WHERE user_id = $1`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.Status, &s.SegmentType, &s.SegmentIndex, &s.PomodoroCount,
		&s.StartedAt, &s.ElapsedSeconds, &s.TotalElapsed, &s.LastWaterAt, &s.LastStretchAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

// Upsert creates or updates the session state for a user.
func (r *SessionRepo) Upsert(ctx context.Context, userID string, s *models.SessionState) (*models.SessionState, error) {
	s.UserID = userID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sessions (user_id, status, segment_type, segment_index, pomodoro_count,
		                       started_at, elapsed_seconds, total_elapsed, last_water_at, last_stretch_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (user_id) DO UPDATE SET
		   status          = EXCLUDED.status,
		   segment_type    = EXCLUDED.segment_type,
		   segment_index   = EXCLUDED.segment_index,
		   pomodoro_count  = EXCLUDED.pomodoro_count,
		   started_at      = EXCLUDED.started_at,
		   elapsed_seconds = EXCLUDED.elapsed_seconds,
		   total_elapsed   = EXCLUDED.total_elapsed,
		   last_water_at   = EXCLUDED.last_water_at,
		   last_stretch_at = EXCLUDED.last_stretch_at
		 RETURNING id, user_id, status, segment_type, segment_index, pomodoro_count,
		           started_at, elapsed_seconds, total_elapsed, last_water_at, last_stretch_at`,
		userID, s.Status, s.SegmentType, s.SegmentIndex, s.PomodoroCount,
		s.StartedAt, s.ElapsedSeconds, s.TotalElapsed, s.LastWaterAt, s.LastStretchAt,
	).Scan(&s.ID, &s.UserID, &s.Status, &s.SegmentType, &s.SegmentIndex, &s.PomodoroCount,
		&s.StartedAt, &s.ElapsedSeconds, &s.TotalElapsed, &s.LastWaterAt, &s.LastStretchAt)

	if err != nil {
		return nil, fmt.Errorf("upsert session: %w", err)
	}
	return s, nil
}
