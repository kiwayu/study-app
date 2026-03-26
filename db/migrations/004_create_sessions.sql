CREATE TABLE IF NOT EXISTS sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'idle',
    segment_type    TEXT NOT NULL DEFAULT 'focus',
    segment_index   INTEGER NOT NULL DEFAULT 0,
    pomodoro_count  INTEGER NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ,
    elapsed_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_elapsed   DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_water_at   DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_stretch_at DOUBLE PRECISION NOT NULL DEFAULT 0
);
