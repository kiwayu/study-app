package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler serves the /health endpoint.
type HealthHandler struct {
	pool *pgxpool.Pool
}

// NewHealthHandler creates a HealthHandler. The pool may be nil if no database
// connectivity check is desired.
func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Check returns {"status":"ok","timestamp":"..."} with 200,
// or {"status":"error",...} with 503 if the database is unreachable.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	code := http.StatusOK

	if h.pool != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.pool.Ping(ctx); err != nil {
			status = "error"
			code = http.StatusServiceUnavailable
		}
	}

	writeJSON(w, code, map[string]string{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
