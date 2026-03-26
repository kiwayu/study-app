package models

import "time"

// User represents an authenticated user.
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatarUrl"`
	Provider   string    `json:"provider"`   // "google" or "github"
	ProviderID string    `json:"providerId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Task represents a study task / pomodoro item.
type Task struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"userId"`
	Title              string     `json:"title"`
	EstimatedPomodoros int        `json:"estimatedPomodoros"`
	CompletedPomodoros int        `json:"completedPomodoros"`
	Priority           string     `json:"priority"` // "high", "medium", "low"
	Category           string     `json:"category"`
	Completed          bool       `json:"completed"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	Order              int        `json:"order"`
	SegmentMinutes     int        `json:"segmentMinutes"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// Settings holds user-specific timer configuration.
type Settings struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	PomodoroDuration int    `json:"pomodoroDuration"`
	ShortBreak       int    `json:"shortBreak"`
	LongBreak        int    `json:"longBreak"`
	WaterInterval    int    `json:"waterInterval"`
	StretchInterval  int    `json:"stretchInterval"`
}

// SessionState holds the live pomodoro session for a user.
type SessionState struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	Status         string     `json:"status"`      // "idle", "running", "paused"
	SegmentType    string     `json:"segmentType"` // "focus", "short_break", "long_break"
	SegmentIndex   int        `json:"segmentIndex"`
	PomodoroCount  int        `json:"pomodoroCount"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	ElapsedSeconds float64    `json:"elapsedSeconds"`
	TotalElapsed   float64    `json:"totalElapsed"`
	LastWaterAt    float64    `json:"lastWaterAt"`
	LastStretchAt  float64    `json:"lastStretchAt"`
}

// Note represents a daily session note for a user.
type Note struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	Date   string `json:"date"` // "YYYY-MM-DD"
	Text   string `json:"text"`
}

// RefreshToken stores a hashed refresh token for a user.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}
