package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"studysession/auth"
	"studysession/db"
)

var dateRe = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

// NoteHandler provides HTTP handlers for daily notes backed by PostgreSQL.
type NoteHandler struct {
	repo *db.NoteRepo
}

// NewNoteHandler creates a NoteHandler.
func NewNoteHandler(repo *db.NoteRepo) *NoteHandler {
	return &NoteHandler{repo: repo}
}

// Get returns the note for a given date.
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	date := r.PathValue("date")
	if !dateRe.MatchString(date) {
		writeError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	note, err := h.repo.Get(r.Context(), userID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	text := ""
	if note != nil {
		text = note.Text
	}
	writeJSON(w, http.StatusOK, map[string]string{"date": date, "text": text})
}

// Upsert creates or updates the note for a given date.
func (h *NoteHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	date := r.PathValue("date")
	if !dateRe.MatchString(date) {
		writeError(w, http.StatusBadRequest, "invalid date format")
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.Text) > 2000 {
		writeError(w, http.StatusBadRequest, "text must be 2000 characters or fewer")
		return
	}

	_, err := h.repo.Upsert(r.Context(), userID, date, body.Text)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save note")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"date": date, "text": body.Text})
}
