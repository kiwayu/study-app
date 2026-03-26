package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studysession/models"
)

// TaskRepo handles task persistence.
type TaskRepo struct {
	pool *pgxpool.Pool
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

// List returns all tasks for a user, ordered by "order" ascending.
func (r *TaskRepo) List(ctx context.Context, userID string) ([]models.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, estimated_pomodoros, completed_pomodoros,
		        priority, category, completed, completed_at, "order",
		        segment_minutes, created_at, updated_at
		 FROM tasks WHERE user_id = $1 ORDER BY "order" ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.EstimatedPomodoros, &t.CompletedPomodoros,
			&t.Priority, &t.Category, &t.Completed, &t.CompletedAt, &t.Order,
			&t.SegmentMinutes, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	return tasks, rows.Err()
}

// Create inserts a new task for the given user.
func (r *TaskRepo) Create(ctx context.Context, userID string, t *models.Task) (*models.Task, error) {
	t.UserID = userID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, title, estimated_pomodoros, completed_pomodoros,
		                    priority, category, completed, completed_at, "order", segment_minutes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at, updated_at`,
		userID, t.Title, t.EstimatedPomodoros, t.CompletedPomodoros,
		t.Priority, t.Category, t.Completed, t.CompletedAt, t.Order, t.SegmentMinutes,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

// Update modifies an existing task. Only the fields present in the map are updated.
// The map keys should match the DB column names.
func (r *TaskRepo) Update(ctx context.Context, userID, taskID string, fields map[string]any) (*models.Task, error) {
	// Build SET clause dynamically.
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	allowed := map[string]bool{
		"title": true, "estimated_pomodoros": true, "completed_pomodoros": true,
		"priority": true, "category": true, "completed": true, "completed_at": true,
		"order": true, "segment_minutes": true,
	}

	for col, val := range fields {
		if !allowed[col] {
			continue
		}
		colName := col
		if col == "order" {
			colName = `"order"`
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", colName, argIdx))
		args = append(args, val)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	// Always bump updated_at.
	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(
		`UPDATE tasks SET %s WHERE id = $%d AND user_id = $%d
		 RETURNING id, user_id, title, estimated_pomodoros, completed_pomodoros,
		           priority, category, completed, completed_at, "order",
		           segment_minutes, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx, argIdx+1,
	)
	args = append(args, taskID, userID)

	var t models.Task
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.UserID, &t.Title, &t.EstimatedPomodoros, &t.CompletedPomodoros,
		&t.Priority, &t.Category, &t.Completed, &t.CompletedAt, &t.Order,
		&t.SegmentMinutes, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return &t, nil
}

// Delete removes a task by ID, scoped to the given user.
func (r *TaskRepo) Delete(ctx context.Context, userID, taskID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM tasks WHERE id = $1 AND user_id = $2`,
		taskID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

