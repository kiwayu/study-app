package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	webview "github.com/jchv/go-webview2"
)


//go:embed static
var staticFiles embed.FS

// ---- Structs ----------------------------------------------------------------

type Task struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	EstimatedPomodoros int    `json:"estimatedPomodoros"`
	CompletedPomodoros int    `json:"completedPomodoros"`
	Priority           string `json:"priority"` // "high" | "medium" | "low"
	Category           string `json:"category"`
	Completed          bool   `json:"completed"`
	Order              int    `json:"order"`
}

type Settings struct {
	PomodoroDuration int `json:"pomodoroDuration"` // minutes
	ShortBreak       int `json:"shortBreak"`        // minutes
	LongBreak        int `json:"longBreak"`          // minutes
	WaterInterval    int `json:"waterInterval"`      // minutes
	StretchInterval  int `json:"stretchInterval"`    // minutes
}

type SessionState struct {
	Status         string     `json:"status"`         // "idle" | "running" | "paused"
	SegmentType    string     `json:"segmentType"`    // "focus" | "short_break" | "long_break"
	SegmentIndex   int        `json:"segmentIndex"`   // 0–7, wraps at 8
	PomodoroCount  int        `json:"pomodoroCount"`
	StartedAt      *time.Time `json:"startedAt"`
	ElapsedSeconds float64    `json:"elapsedSeconds"`
	TotalElapsed   float64    `json:"totalElapsed"`
	LastWaterAt    float64    `json:"lastWaterAt"`
	LastStretchAt  float64    `json:"lastStretchAt"`
}

type AppState struct {
	Tasks    []Task       `json:"tasks"`
	Settings Settings     `json:"settings"`
	Session  SessionState `json:"session"`
}

// ---- Request types ----------------------------------------------------------

type CreateTaskRequest struct {
	Title              string `json:"title"`
	EstimatedPomodoros int    `json:"estimatedPomodoros"`
	Priority           string `json:"priority"`
	Category           string `json:"category"`
}

type UpdateTaskRequest struct {
	Title              *string `json:"title"`
	EstimatedPomodoros *int    `json:"estimatedPomodoros"`
	CompletedPomodoros *int    `json:"completedPomodoros"`
	Priority           *string `json:"priority"`
	Category           *string `json:"category"`
	Completed          *bool   `json:"completed"`
	Order              *int    `json:"order"`
}

type StartRequest struct {
	SegmentType   string `json:"segmentType"`
	SegmentIndex  int    `json:"segmentIndex"`
	PomodoroCount int    `json:"pomodoroCount"`
}

type TotalsRequest struct {
	TotalElapsed  float64 `json:"totalElapsed"`
	LastWaterAt   float64 `json:"lastWaterAt"`
	LastStretchAt float64 `json:"lastStretchAt"`
}

// ---- Global store -----------------------------------------------------------

var (
	mu        sync.RWMutex
	store     AppState
	dataDir   = "data"
	stateFile = "data/state.json"
)

func defaultState() AppState {
	return AppState{
		Tasks: []Task{},
		Settings: Settings{
			PomodoroDuration: 25,
			ShortBreak:       5,
			LongBreak:        15,
			WaterInterval:    45,
			StretchInterval:  60,
		},
		Session: SessionState{
			Status:      "idle",
			SegmentType: "focus",
		},
	}
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func loadState() {
	store = defaultState()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return // first run — use defaults
	}
	if err := json.Unmarshal(data, &store); err != nil {
		log.Printf("loadState: corrupt state.json, using defaults: %v", err)
		store = defaultState()
		return
	}
	if store.Tasks == nil {
		store.Tasks = []Task{}
	}
}

// persistState must be called with mu write-locked. Disk I/O inside the lock
// is acceptable for this single-user localhost app (sub-ms on any local FS).
func persistState() {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		log.Printf("persistState marshal: %v", err)
		return
	}
	tmp := stateFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		log.Printf("persistState write: %v", err)
		return
	}
	if err := os.Rename(tmp, stateFile); err != nil {
		log.Printf("persistState rename: %v", err)
	}
}

// ---- Response helpers -------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- CORS middleware --------------------------------------------------------

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---- Task handlers ----------------------------------------------------------

func listTasks(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	tasks := make([]Task, len(store.Tasks))
	copy(tasks, store.Tasks)
	mu.RUnlock()
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })
	writeJSON(w, http.StatusOK, tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	maxOrder := 0
	for _, t := range store.Tasks {
		if t.Order >= maxOrder {
			maxOrder = t.Order + 1
		}
	}
	task := Task{
		ID:                 generateID(),
		Title:              req.Title,
		EstimatedPomodoros: req.EstimatedPomodoros,
		Priority:           req.Priority,
		Category:           req.Category,
		Order:              maxOrder,
	}
	store.Tasks = append(store.Tasks, task)
	persistState()
	mu.Unlock()

	writeJSON(w, http.StatusCreated, task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mu.Lock()
	defer mu.Unlock()
	for i, t := range store.Tasks {
		if t.ID != id {
			continue
		}
		if req.Title != nil {
			store.Tasks[i].Title = *req.Title
		}
		if req.EstimatedPomodoros != nil {
			store.Tasks[i].EstimatedPomodoros = *req.EstimatedPomodoros
		}
		if req.CompletedPomodoros != nil {
			store.Tasks[i].CompletedPomodoros = *req.CompletedPomodoros
		}
		if req.Priority != nil {
			store.Tasks[i].Priority = *req.Priority
		}
		if req.Category != nil {
			store.Tasks[i].Category = *req.Category
		}
		if req.Completed != nil {
			store.Tasks[i].Completed = *req.Completed
		}
		if req.Order != nil {
			store.Tasks[i].Order = *req.Order
		}
		persistState()
		writeJSON(w, http.StatusOK, store.Tasks[i])
		return
	}
	writeError(w, http.StatusNotFound, "task not found")
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	mu.Lock()
	defer mu.Unlock()
	idx := -1
	for i, t := range store.Tasks {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	store.Tasks = append(store.Tasks[:idx], store.Tasks[idx+1:]...)
	sort.Slice(store.Tasks, func(i, j int) bool { return store.Tasks[i].Order < store.Tasks[j].Order })
	for i := range store.Tasks {
		store.Tasks[i].Order = i
	}
	persistState()
	w.WriteHeader(http.StatusNoContent)
}
func getSettings(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	s := store.Settings
	mu.RUnlock()
	writeJSON(w, http.StatusOK, s)
}

func putSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if s.PomodoroDuration < 1 || s.ShortBreak < 1 || s.LongBreak < 1 ||
		s.WaterInterval < 1 || s.StretchInterval < 1 {
		writeError(w, http.StatusBadRequest, "all durations must be at least 1 minute")
		return
	}
	mu.Lock()
	store.Settings = s
	persistState()
	mu.Unlock()
	writeJSON(w, http.StatusOK, s)
}

func getSession(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	sess := store.Session
	mu.RUnlock()
	writeJSON(w, http.StatusOK, sess)
}

func startSession(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	mu.Lock()
	// If currently running, bank elapsed time into TotalElapsed before resetting
	if store.Session.Status == "running" && store.Session.StartedAt != nil {
		store.Session.TotalElapsed += time.Since(*store.Session.StartedAt).Seconds()
	}
	store.Session.Status = "running"
	store.Session.SegmentType = req.SegmentType
	store.Session.SegmentIndex = req.SegmentIndex
	store.Session.PomodoroCount = req.PomodoroCount
	store.Session.StartedAt = &now
	store.Session.ElapsedSeconds = 0
	persistState()
	sess := store.Session
	mu.Unlock()
	writeJSON(w, http.StatusOK, sess)
}

func pauseSession(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	if store.Session.Status == "running" && store.Session.StartedAt != nil {
		elapsed := time.Since(*store.Session.StartedAt).Seconds()
		store.Session.ElapsedSeconds += elapsed
		store.Session.TotalElapsed += elapsed
	}
	store.Session.Status = "paused"
	store.Session.StartedAt = nil
	persistState()
	sess := store.Session
	mu.Unlock()
	writeJSON(w, http.StatusOK, sess)
}

func stopSession(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	preserved := struct{ total, water, stretch float64 }{
		store.Session.TotalElapsed,
		store.Session.LastWaterAt,
		store.Session.LastStretchAt,
	}
	store.Session = SessionState{
		Status:        "idle",
		SegmentType:   "focus",
		TotalElapsed:  preserved.total,
		LastWaterAt:   preserved.water,
		LastStretchAt: preserved.stretch,
	}
	persistState()
	sess := store.Session
	mu.Unlock()
	writeJSON(w, http.StatusOK, sess)
}

func updateTotals(w http.ResponseWriter, r *http.Request) {
	var req TotalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	mu.Lock()
	store.Session.TotalElapsed = req.TotalElapsed
	store.Session.LastWaterAt = req.LastWaterAt
	store.Session.LastStretchAt = req.LastStretchAt
	persistState()
	sess := store.Session
	mu.Unlock()
	writeJSON(w, http.StatusOK, sess)
}

// ---- Main -------------------------------------------------------------------

func main() {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	loadState()

	mux := http.NewServeMux()

	// Task routes
	mux.HandleFunc("GET /api/tasks", listTasks)
	mux.HandleFunc("POST /api/tasks", createTask)
	mux.HandleFunc("PUT /api/tasks/{id}", updateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", deleteTask)

	// Settings routes
	mux.HandleFunc("GET /api/settings", getSettings)
	mux.HandleFunc("PUT /api/settings", putSettings)

	// Session routes
	mux.HandleFunc("GET /api/session", getSession)
	mux.HandleFunc("POST /api/session/start", startSession)
	mux.HandleFunc("POST /api/session/pause", pauseSession)
	mux.HandleFunc("POST /api/session/stop", stopSession)
	mux.HandleFunc("PUT /api/session/totals", updateTotals)

	// Static files (catch-all)
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("sub static FS: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	// Pick a random free port on localhost
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Start HTTP server in background
	go func() {
		log.Printf("Listening on %s", url)
		if err := http.Serve(ln, corsMiddleware(mux)); err != nil {
			log.Printf("server: %v", err)
		}
	}()

	// Give the server a moment to bind
	time.Sleep(100 * time.Millisecond)

	// Open native desktop window
	w := webview.NewWithOptions(webview.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview.WindowOptions{
			Title:  "Study Session",
			Width:  393,
			Height: 852,
		},
	})
	defer w.Destroy()

	// Get the native window handle
	hwnd := uintptr(w.Window())

	// Load SetWindowPos from user32
	user32 := syscall.NewLazyDLL("user32.dll")
	setWindowPos := user32.NewProc("SetWindowPos")

	const (
		hwndTopmost    = ^uintptr(0)     // (HWND)-1
		hwndNotTopmost = ^uintptr(0) - 1 // (HWND)-2
		swpNoMove      = 0x0002
		swpNoSize      = 0x0001
	)

	alwaysOnTop := false
	w.Bind("goToggleAlwaysOnTop", func() error {
		alwaysOnTop = !alwaysOnTop
		insertAfter := hwndNotTopmost
		if alwaysOnTop {
			insertAfter = hwndTopmost
		}
		setWindowPos.Call(hwnd, insertAfter, 0, 0, 0, 0, swpNoMove|swpNoSize)
		return nil
	})

	w.SetSize(393, 852, webview.HintNone)
	w.Navigate(url)
	w.Run()
}
