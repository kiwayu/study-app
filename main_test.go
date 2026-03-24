package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "studysession-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)
	dataDir = tmp
	stateFile = filepath.Join(tmp, "state.json")
	os.Exit(m.Run())
}

func resetStore() {
	store = defaultState()
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func decodeBody[T any](t *testing.T, body *bytes.Buffer) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(body).Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return v
}

func TestListTasksEmpty(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()
	listTasks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	tasks := decodeBody[[]Task](t, w.Body)
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestCreateAndListTask(t *testing.T) {
	resetStore()
	body := jsonBody(CreateTaskRequest{
		Title: "Write tests", EstimatedPomodoros: 2, Priority: "high",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", body)
	w := httptest.NewRecorder()
	createTask(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	task := decodeBody[Task](t, w.Body)
	if task.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if task.Title != "Write tests" {
		t.Fatalf("expected 'Write tests', got %q", task.Title)
	}
	if task.Order != 0 {
		t.Fatalf("expected order 0, got %d", task.Order)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	resetStore()
	cases := []struct {
		name string
		body CreateTaskRequest
		code int
	}{
		{"missing title", CreateTaskRequest{EstimatedPomodoros: 1, Priority: "high"}, 400},
		{"zero pomodoros", CreateTaskRequest{Title: "x", EstimatedPomodoros: 0, Priority: "high"}, 400},
		{"bad priority", CreateTaskRequest{Title: "x", EstimatedPomodoros: 1, Priority: "critical"}, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/tasks", jsonBody(tc.body))
			w := httptest.NewRecorder()
			createTask(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks",
		jsonBody(CreateTaskRequest{Title: "Old", EstimatedPomodoros: 1, Priority: "low"}))
	w := httptest.NewRecorder()
	createTask(w, req)
	task := decodeBody[Task](t, w.Body)

	newTitle := "New"
	req2 := httptest.NewRequest(http.MethodPut, "/api/tasks/"+task.ID,
		jsonBody(UpdateTaskRequest{Title: &newTitle}))
	req2.SetPathValue("id", task.ID)
	w2 := httptest.NewRecorder()
	updateTask(w2, req2)

	if w2.Code != 200 {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	updated := decodeBody[Task](t, w2.Body)
	if updated.Title != "New" {
		t.Fatalf("expected 'New', got %q", updated.Title)
	}
}

func TestDeleteTask(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks",
		jsonBody(CreateTaskRequest{Title: "Del me", EstimatedPomodoros: 1, Priority: "low"}))
	w := httptest.NewRecorder()
	createTask(w, req)
	task := decodeBody[Task](t, w.Body)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+task.ID, nil)
	req2.SetPathValue("id", task.ID)
	w2 := httptest.NewRecorder()
	deleteTask(w2, req2)

	if w2.Code != 204 {
		t.Fatalf("expected 204, got %d", w2.Code)
	}
	if len(store.Tasks) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(store.Tasks))
	}
}

func TestGetSettings(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	getSettings(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	s := decodeBody[Settings](t, w.Body)
	if s.PomodoroDuration != 25 {
		t.Fatalf("expected default 25, got %d", s.PomodoroDuration)
	}
}

func TestPutSettings(t *testing.T) {
	resetStore()
	newSettings := Settings{
		PomodoroDuration: 30, ShortBreak: 6, LongBreak: 20,
		WaterInterval: 50, StretchInterval: 70,
	}
	req := httptest.NewRequest(http.MethodPut, "/api/settings", jsonBody(newSettings))
	w := httptest.NewRecorder()
	putSettings(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	s := decodeBody[Settings](t, w.Body)
	if s.PomodoroDuration != 30 {
		t.Fatalf("expected 30, got %d", s.PomodoroDuration)
	}
}

func TestSessionLifecycle(t *testing.T) {
	resetStore()

	// Start
	req := httptest.NewRequest(http.MethodPost, "/api/session/start",
		jsonBody(StartRequest{SegmentType: "focus", SegmentIndex: 0, PomodoroCount: 0}))
	w := httptest.NewRecorder()
	startSession(w, req)
	if w.Code != 200 {
		t.Fatalf("start: expected 200, got %d", w.Code)
	}
	sess := decodeBody[SessionState](t, w.Body)
	if sess.Status != "running" {
		t.Fatalf("expected running, got %q", sess.Status)
	}
	if sess.StartedAt == nil {
		t.Fatal("startedAt should not be nil")
	}

	// Pause
	req2 := httptest.NewRequest(http.MethodPost, "/api/session/pause", nil)
	w2 := httptest.NewRecorder()
	pauseSession(w2, req2)
	if w2.Code != 200 {
		t.Fatalf("pause: expected 200, got %d", w2.Code)
	}
	sess2 := decodeBody[SessionState](t, w2.Body)
	if sess2.Status != "paused" {
		t.Fatalf("expected paused, got %q", sess2.Status)
	}
	if sess2.StartedAt != nil {
		t.Fatal("startedAt should be nil after pause")
	}

	// Stop
	req3 := httptest.NewRequest(http.MethodPost, "/api/session/stop", nil)
	w3 := httptest.NewRecorder()
	stopSession(w3, req3)
	if w3.Code != 200 {
		t.Fatalf("stop: expected 200, got %d", w3.Code)
	}
	sess3 := decodeBody[SessionState](t, w3.Body)
	if sess3.Status != "idle" {
		t.Fatalf("expected idle, got %q", sess3.Status)
	}
	if sess3.SegmentIndex != 0 || sess3.PomodoroCount != 0 {
		t.Fatalf("stop should reset index/count, got index=%d count=%d", sess3.SegmentIndex, sess3.PomodoroCount)
	}
}

func TestUpdateTotals(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodPut, "/api/session/totals",
		jsonBody(TotalsRequest{TotalElapsed: 3600, LastWaterAt: 2700, LastStretchAt: 3600}))
	w := httptest.NewRecorder()
	updateTotals(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	sess := decodeBody[SessionState](t, w.Body)
	if sess.TotalElapsed != 3600 {
		t.Fatalf("expected 3600, got %f", sess.TotalElapsed)
	}
}
