package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// SettingsRepo handles user settings persistence.
type SettingsRepo struct {
	pool *pgxpool.Pool
}

// NewSettingsRepo creates a new SettingsRepo.
func NewSettingsRepo(pool *pgxpool.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

// Get retrieves settings for a user. Returns nil, nil if not found.
func (r *SettingsRepo) Get(ctx context.Context, userID string) (*models.Settings, error) {
	s := &models.Settings{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, pomodoro_duration, short_break, long_break, water_interval, stretch_interval
		 FROM settings WHERE user_id = $1`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.PomodoroDuration, &s.ShortBreak, &s.LongBreak, &s.WaterInterval, &s.StretchInterval)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return s, nil
}

// Upsert creates or updates settings for a user.
func (r *SettingsRepo) Upsert(ctx context.Context, userID string, s *models.Settings) (*models.Settings, error) {
	s.UserID = userID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO settings (user_id, pomodoro_duration, short_break, long_break, water_interval, stretch_interval)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id) DO UPDATE SET
		   pomodoro_duration = EXCLUDED.pomodoro_duration,
		   short_break       = EXCLUDED.short_break,
		   long_break        = EXCLUDED.long_break,
		   water_interval    = EXCLUDED.water_interval,
		   stretch_interval  = EXCLUDED.stretch_interval
		 RETURNING id, user_id, pomodoro_duration, short_break, long_break, water_interval, stretch_interval`,
		userID, s.PomodoroDuration, s.ShortBreak, s.LongBreak, s.WaterInterval, s.StretchInterval,
	).Scan(&s.ID, &s.UserID, &s.PomodoroDuration, &s.ShortBreak, &s.LongBreak, &s.WaterInterval, &s.StretchInterval)

	if err != nil {
		return nil, fmt.Errorf("upsert settings: %w", err)
	}
	return s, nil
}
