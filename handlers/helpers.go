package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// CreateTaskRequest is the body for POST /api/tasks.
type CreateTaskRequest struct {
	Title              string `json:"title"`
	EstimatedPomodoros int    `json:"estimatedPomodoros"`
	Priority           string `json:"priority"`
	Category           string `json:"category"`
	SegmentMinutes     int    `json:"segmentMinutes"`
}

// UpdateTaskRequest is the body for PUT /api/tasks/{id}.
// Pointer fields allow distinguishing "not sent" from zero-value.
type UpdateTaskRequest struct {
	Title              *string `json:"title"`
	EstimatedPomodoros *int    `json:"estimatedPomodoros"`
	CompletedPomodoros *int    `json:"completedPomodoros"`
	Priority           *string `json:"priority"`
	Category           *string `json:"category"`
	Completed          *bool   `json:"completed"`
	Order              *int    `json:"order"`
	SegmentMinutes     *int    `json:"segmentMinutes"`
}

// StartRequest is the body for POST /api/session/start.
type StartRequest struct {
	SegmentType   string `json:"segmentType"`
	SegmentIndex  int    `json:"segmentIndex"`
	PomodoroCount int    `json:"pomodoroCount"`
}

// TotalsRequest is the body for PUT /api/session/totals.
type TotalsRequest struct {
	TotalElapsed  float64 `json:"totalElapsed"`
	LastWaterAt   float64 `json:"lastWaterAt"`
	LastStretchAt float64 `json:"lastStretchAt"`
}
