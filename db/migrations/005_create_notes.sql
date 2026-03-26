CREATE TABLE IF NOT EXISTS notes (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date    TEXT NOT NULL,
    text    TEXT NOT NULL DEFAULT '',
    UNIQUE (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
