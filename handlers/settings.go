package handlers

import (
	"encoding/json"
	"net/http"

	"studysession/auth"
	"studysession/db"
	"studysession/models"
)

// SettingsHandler provides HTTP handlers for user settings backed by PostgreSQL.
type SettingsHandler struct {
	repo *db.SettingsRepo
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(repo *db.SettingsRepo) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

// Get returns settings for the authenticated user, falling back to defaults.
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	s, err := h.repo.Get(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings")
		return
	}
	if s == nil {
		// Return defaults.
		s = &models.Settings{
			PomodoroDuration: 25,
			ShortBreak:       5,
			LongBreak:        15,
			WaterInterval:    45,
			StretchInterval:  60,
		}
	}
	writeJSON(w, http.StatusOK, s)
}

// Put updates settings for the authenticated user.
func (h *SettingsHandler) Put(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var s models.Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if s.PomodoroDuration < 1 || s.PomodoroDuration > 120 ||
		s.ShortBreak < 1 || s.ShortBreak > 60 ||
		s.LongBreak < 1 || s.LongBreak > 120 ||
		s.WaterInterval < 1 || s.WaterInterval > 480 ||
		s.StretchInterval < 1 || s.StretchInterval > 480 {
		writeError(w, http.StatusBadRequest, "duration values out of allowed range")
		return
	}

	saved, err := h.repo.Upsert(r.Context(), userID, &s)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}
