package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"studysession/auth"
	"studysession/db"
	"studysession/models"
)

// SessionHandler provides HTTP handlers for the pomodoro session backed by PostgreSQL.
type SessionHandler struct {
	repo *db.SessionRepo
}

// NewSessionHandler creates a SessionHandler.
func NewSessionHandler(repo *db.SessionRepo) *SessionHandler {
	return &SessionHandler{repo: repo}
}

// defaultSession returns the initial idle session state.
func defaultSession() *models.SessionState {
	return &models.SessionState{
		Status:      "idle",
		SegmentType: "focus",
	}
}

// Get returns the current session state for the authenticated user.
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	sess, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if sess == nil {
		sess = defaultSession()
	}
	writeJSON(w, http.StatusOK, sess)
}

// Start begins or restarts a session segment.
func (h *SessionHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	validSegTypes := map[string]bool{"focus": true, "short_break": true, "long_break": true}
	if !validSegTypes[req.SegmentType] {
		writeError(w, http.StatusBadRequest, "segmentType must be focus, short_break, or long_break")
		return
	}
	if req.SegmentIndex < 0 || req.SegmentIndex > 7 {
		writeError(w, http.StatusBadRequest, "segmentIndex must be between 0 and 7")
		return
	}
	if req.PomodoroCount < 0 {
		writeError(w, http.StatusBadRequest, "pomodoroCount cannot be negative")
		return
	}

	// Get existing session to bank elapsed time.
	sess, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if sess == nil {
		sess = defaultSession()
	}

	// If currently running, bank elapsed time.
	if sess.Status == "running" && sess.StartedAt != nil {
		sess.TotalElapsed += time.Since(*sess.StartedAt).Seconds()
	}

	now := time.Now()
	sess.Status = "running"
	sess.SegmentType = req.SegmentType
	sess.SegmentIndex = req.SegmentIndex
	sess.PomodoroCount = req.PomodoroCount
	sess.StartedAt = &now
	sess.ElapsedSeconds = 0

	saved, err := h.repo.Upsert(r.Context(), userID, sess)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// Pause pauses the running session.
func (h *SessionHandler) Pause(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	sess, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if sess == nil {
		sess = defaultSession()
	}

	if sess.Status == "running" && sess.StartedAt != nil {
		elapsed := time.Since(*sess.StartedAt).Seconds()
		sess.ElapsedSeconds += elapsed
		sess.TotalElapsed += elapsed
	}
	sess.Status = "paused"
	sess.StartedAt = nil

	saved, err := h.repo.Upsert(r.Context(), userID, sess)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// Stop resets the session to idle, preserving totals.
func (h *SessionHandler) Stop(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	sess, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if sess == nil {
		sess = defaultSession()
	}

	preserved := struct{ total, water, stretch float64 }{
		sess.TotalElapsed,
		sess.LastWaterAt,
		sess.LastStretchAt,
	}

	*sess = models.SessionState{
		Status:        "idle",
		SegmentType:   "focus",
		TotalElapsed:  preserved.total,
		LastWaterAt:   preserved.water,
		LastStretchAt: preserved.stretch,
	}

	saved, err := h.repo.Upsert(r.Context(), userID, sess)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// UpdateTotals sets the session total counters.
func (h *SessionHandler) UpdateTotals(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req TotalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sess, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if sess == nil {
		sess = defaultSession()
	}

	sess.TotalElapsed = req.TotalElapsed
	sess.LastWaterAt = req.LastWaterAt
	sess.LastStretchAt = req.LastStretchAt

	saved, err := h.repo.Upsert(r.Context(), userID, sess)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}
