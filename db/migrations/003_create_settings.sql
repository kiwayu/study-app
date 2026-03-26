CREATE TABLE IF NOT EXISTS settings (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    pomodoro_duration INTEGER NOT NULL DEFAULT 25,
    short_break       INTEGER NOT NULL DEFAULT 5,
    long_break        INTEGER NOT NULL DEFAULT 15,
    water_interval    INTEGER NOT NULL DEFAULT 45,
    stretch_interval  INTEGER NOT NULL DEFAULT 60
);
