package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"studysession/auth"
	"studysession/db"
	"studysession/models"
)

// TaskHandler provides HTTP handlers for task CRUD backed by PostgreSQL.
type TaskHandler struct {
	repo *db.TaskRepo
}

// NewTaskHandler creates a TaskHandler.
func NewTaskHandler(repo *db.TaskRepo) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// List returns all tasks for the authenticated user.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	tasks, err := h.repo.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// Create adds a new task for the authenticated user.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.EstimatedPomodoros < 1 {
		writeError(w, http.StatusBadRequest, "estimatedPomodoros must be at least 1")
		return
	}
	if req.Priority != "high" && req.Priority != "medium" && req.Priority != "low" {
		writeError(w, http.StatusBadRequest, "priority must be high, medium, or low")
		return
	}
	if len(req.Title) > 200 {
		writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
		return
	}
	if len(req.Category) > 100 {
		writeError(w, http.StatusBadRequest, "category must be 100 characters or fewer")
		return
	}

	task := &models.Task{
		Title:              req.Title,
		EstimatedPomodoros: req.EstimatedPomodoros,
		Priority:           req.Priority,
		Category:           req.Category,
		SegmentMinutes:     req.SegmentMinutes,
	}

	created, err := h.repo.Create(r.Context(), userID, task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Update modifies an existing task for the authenticated user.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := make(map[string]any)

	if req.Title != nil {
		if len(*req.Title) > 200 {
			writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
			return
		}
		fields["title"] = *req.Title
	}
	if req.Category != nil {
		if len(*req.Category) > 100 {
			writeError(w, http.StatusBadRequest, "category must be 100 characters or fewer")
			return
		}
		fields["category"] = *req.Category
	}
	if req.EstimatedPomodoros != nil {
		if *req.EstimatedPomodoros < 1 {
			writeError(w, http.StatusBadRequest, "estimatedPomodoros must be at least 1")
			return
		}
		fields["estimated_pomodoros"] = *req.EstimatedPomodoros
	}
	if req.CompletedPomodoros != nil {
		if *req.CompletedPomodoros < 0 {
			writeError(w, http.StatusBadRequest, "completedPomodoros cannot be negative")
			return
		}
		fields["completed_pomodoros"] = *req.CompletedPomodoros
	}
	if req.Priority != nil {
		if *req.Priority != "high" && *req.Priority != "medium" && *req.Priority != "low" {
			writeError(w, http.StatusBadRequest, "priority must be high, medium, or low")
			return
		}
		fields["priority"] = *req.Priority
	}
	if req.Completed != nil {
		fields["completed"] = *req.Completed
		if *req.Completed {
			// When marking complete, set completed_at to now.
			now := time.Now()
			fields["completed_at"] = &now
		} else {
			// When un-completing, clear completed_at.
			fields["completed_at"] = nil
		}
	}
	if req.Order != nil {
		if *req.Order < 0 {
			writeError(w, http.StatusBadRequest, "order cannot be negative")
			return
		}
		fields["order"] = *req.Order
	}
	if req.SegmentMinutes != nil {
		fields["segment_minutes"] = *req.SegmentMinutes
	}

	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	updated, err := h.repo.Update(r.Context(), userID, id, fields)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete removes a task for the authenticated user.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.repo.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Completions returns daily completion counts for the authenticated user.
func (h *TaskHandler) Completions(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	tasks, err := h.repo.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}

	counts := make(map[string]int)
	for _, t := range tasks {
		if t.Completed && t.CompletedAt != nil {
			day := t.CompletedAt.Local().Format("2006-01-02")
			counts[day]++
		}
	}
	writeJSON(w, http.StatusOK, counts)
}

// Estimation returns estimation accuracy data for completed tasks.
func (h *TaskHandler) Estimation(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	tasks, err := h.repo.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}

	type EstItem struct {
		Title       string `json:"title"`
		Estimated   int    `json:"estimated"`
		Actual      int    `json:"actual"`
		CompletedAt string `json:"completedAt"`
	}

	var result []EstItem
	for _, t := range tasks {
		if !t.Completed || t.CompletedAt == nil {
			continue
		}
		title := t.Title
		if len([]rune(title)) > 40 {
			runes := []rune(title)
			title = string(runes[:40])
		}
		result = append(result, EstItem{
			Title:       title,
			Estimated:   t.EstimatedPomodoros,
			Actual:      t.CompletedPomodoros,
			CompletedAt: t.CompletedAt.Local().Format("2006-01-02"),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CompletedAt > result[j].CompletedAt
	})

	if len(result) > 50 {
		result = result[:50]
	}
	if result == nil {
		result = []EstItem{}
	}

	writeJSON(w, http.StatusOK, result)
}
