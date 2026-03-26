package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	webview "github.com/jchv/go-webview2"

	"studysession/auth"
	"studysession/config"
	"studysession/db"
	"studysession/handlers"
	"studysession/middleware"
)

//go:embed static
var staticFiles embed.FS

// ---- Desktop-mode structs (JSON persistence) --------------------------------

type Task struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	EstimatedPomodoros int        `json:"estimatedPomodoros"`
	CompletedPomodoros int        `json:"completedPomodoros"`
	Priority           string     `json:"priority"`
	Category           string     `json:"category"`
	Completed          bool       `json:"completed"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	Order              int        `json:"order"`
	SegmentMinutes     int        `json:"segmentMinutes"`
}

type Settings struct {
	PomodoroDuration int `json:"pomodoroDuration"`
	ShortBreak       int `json:"shortBreak"`
	LongBreak        int `json:"longBreak"`
	WaterInterval    int `json:"waterInterval"`
	StretchInterval  int `json:"stretchInterval"`
}

type SessionState struct {
	Status         string     `json:"status"`
	SegmentType    string     `json:"segmentType"`
	SegmentIndex   int        `json:"segmentIndex"`
	PomodoroCount  int        `json:"pomodoroCount"`
	StartedAt      *time.Time `json:"startedAt"`
	ElapsedSeconds float64    `json:"elapsedSeconds"`
	TotalElapsed   float64    `json:"totalElapsed"`
	LastWaterAt    float64    `json:"lastWaterAt"`
	LastStretchAt  float64    `json:"lastStretchAt"`
}

type AppState struct {
	Tasks    []Task            `json:"tasks"`
	Settings Settings          `json:"settings"`
	Session  SessionState      `json:"session"`
	Notes    map[string]string `json:"notes"`
}

// ---- Desktop-mode request types ---------------------------------------------

type CreateTaskRequest struct {
	Title              string `json:"title"`
	EstimatedPomodoros int    `json:"estimatedPomodoros"`
	Priority           string `json:"priority"`
	Category           string `json:"category"`
	SegmentMinutes     int    `json:"segmentMinutes"`
}

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

// ---- Global store (desktop mode) --------------------------------------------

var (
	mu        sync.RWMutex
	store     AppState
	dataDir   string
	stateFile string
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

func hexToCOLORREF(hex string) uint32 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func sanitiseSettings(s *Settings) {
	def := defaultState().Settings
	if s.PomodoroDuration < 1 || s.PomodoroDuration > 120 { s.PomodoroDuration = def.PomodoroDuration }
	if s.ShortBreak < 1 || s.ShortBreak > 60            { s.ShortBreak = def.ShortBreak }
	if s.LongBreak < 1 || s.LongBreak > 120             { s.LongBreak = def.LongBreak }
	if s.WaterInterval < 1 || s.WaterInterval > 480     { s.WaterInterval = def.WaterInterval }
	if s.StretchInterval < 1 || s.StretchInterval > 480 { s.StretchInterval = def.StretchInterval }
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("rand.Read: %v", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func loadState() {
	store = defaultState()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &store); err != nil {
		log.Printf("loadState: corrupt state.json, using defaults: %v", err)
		store = defaultState()
		return
	}
	if store.Tasks == nil {
		store.Tasks = []Task{}
	}
	if store.Notes == nil {
		store.Notes = map[string]string{}
	}
	sanitiseSettings(&store.Settings)
}

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
		os.Remove(tmp)
	}
}

// ---- Response helpers (desktop mode) ----------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---- Desktop-mode CORS middleware -------------------------------------------

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" || strings.HasPrefix(origin, "http://127.0.0.1") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---- Desktop-mode task handlers ---------------------------------------------

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
	if len(req.Title) > 200 {
		writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
		return
	}
	if len(req.Category) > 100 {
		writeError(w, http.StatusBadRequest, "category must be 100 characters or fewer")
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
		SegmentMinutes:     req.SegmentMinutes,
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
			if len(*req.Title) > 200 {
				writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
				return
			}
			store.Tasks[i].Title = *req.Title
		}
		if req.Category != nil {
			if len(*req.Category) > 100 {
				writeError(w, http.StatusBadRequest, "category must be 100 characters or fewer")
				return
			}
			store.Tasks[i].Category = *req.Category
		}
		if req.EstimatedPomodoros != nil {
			if *req.EstimatedPomodoros < 1 {
				writeError(w, http.StatusBadRequest, "estimatedPomodoros must be at least 1")
				return
			}
			store.Tasks[i].EstimatedPomodoros = *req.EstimatedPomodoros
		}
		if req.CompletedPomodoros != nil {
			if *req.CompletedPomodoros < 0 {
				writeError(w, http.StatusBadRequest, "completedPomodoros cannot be negative")
				return
			}
			store.Tasks[i].CompletedPomodoros = *req.CompletedPomodoros
		}
		if req.Priority != nil {
			if *req.Priority != "high" && *req.Priority != "medium" && *req.Priority != "low" {
				writeError(w, http.StatusBadRequest, "priority must be high, medium, or low")
				return
			}
			store.Tasks[i].Priority = *req.Priority
		}
		if req.Completed != nil {
			if *req.Completed && !store.Tasks[i].Completed {
				now := time.Now().Local()
				store.Tasks[i].CompletedAt = &now
			} else if !*req.Completed {
				store.Tasks[i].CompletedAt = nil
			}
			store.Tasks[i].Completed = *req.Completed
		}
		if req.Order != nil {
			if *req.Order < 0 {
				writeError(w, http.StatusBadRequest, "order cannot be negative")
				return
			}
			store.Tasks[i].Order = *req.Order
		}
		if req.SegmentMinutes != nil {
			store.Tasks[i].SegmentMinutes = *req.SegmentMinutes
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

// ---- Desktop-mode settings handlers -----------------------------------------

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
	if s.PomodoroDuration < 1 || s.PomodoroDuration > 120 ||
		s.ShortBreak < 1 || s.ShortBreak > 60 ||
		s.LongBreak < 1 || s.LongBreak > 120 ||
		s.WaterInterval < 1 || s.WaterInterval > 480 ||
		s.StretchInterval < 1 || s.StretchInterval > 480 {
		writeError(w, http.StatusBadRequest, "duration values out of allowed range")
		return
	}
	mu.Lock()
	store.Settings = s
	persistState()
	mu.Unlock()
	writeJSON(w, http.StatusOK, s)
}

// ---- Desktop-mode session handlers ------------------------------------------

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
	now := time.Now().Local()
	mu.Lock()
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

// ---- Desktop-mode stats handlers --------------------------------------------

func getEstimationStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	tasks := make([]Task, len(store.Tasks))
	copy(tasks, store.Tasks)
	mu.RUnlock()

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

func getCompletions(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	tasks := make([]Task, len(store.Tasks))
	copy(tasks, store.Tasks)
	mu.RUnlock()

	counts := make(map[string]int)
	for _, t := range tasks {
		if t.Completed && t.CompletedAt != nil {
			day := t.CompletedAt.Local().Format("2006-01-02")
			counts[day]++
		}
	}
	writeJSON(w, http.StatusOK, counts)
}

// ---- Desktop-mode note handlers ---------------------------------------------

var dateRe = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func getNote(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	date := parts[len(parts)-1]
	if !dateRe.MatchString(date) {
		writeError(w, http.StatusBadRequest, "invalid date format")
		return
	}
	mu.RLock()
	text := store.Notes[date]
	mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]string{"date": date, "text": text})
}

func upsertNote(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	date := parts[len(parts)-1]
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
	mu.Lock()
	store.Notes[date] = body.Text
	persistState()
	mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"date": date, "text": body.Text})
}

// ---- Route registration helpers ---------------------------------------------

// registerDesktopRoutes adds all API routes using the in-memory JSON store.
func registerDesktopRoutes(mux *http.ServeMux) {
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

	// Stats routes
	mux.HandleFunc("GET /api/stats/completions", getCompletions)
	mux.HandleFunc("GET /api/stats/estimation", getEstimationStats)

	// Notes routes
	mux.HandleFunc("/api/notes/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getNote(w, r)
		case http.MethodPut:
			upsertNote(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// ---- Main -------------------------------------------------------------------

func main() {
	desktopMode := flag.Bool("desktop", false, "Run in desktop mode (JSON file, no auth, WebView2 window)")
	portFlag := flag.String("port", "", "Port for web mode (overrides PORT env var)")
	flag.Parse()

	if *desktopMode {
		runDesktopMode()
	} else {
		runWebMode(*portFlag)
	}
}

// runWebMode starts the PostgreSQL-backed, OAuth-authenticated web server.
func runWebMode(portOverride string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if portOverride != "" {
		cfg.Port = portOverride
	}

	// Connect to PostgreSQL.
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations.
	if err := db.RunMigrations(pool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// Create repositories.
	userRepo := db.NewUserRepo(pool)
	taskRepo := db.NewTaskRepo(pool)
	settingsRepo := db.NewSettingsRepo(pool)
	sessionRepo := db.NewSessionRepo(pool)
	noteRepo := db.NewNoteRepo(pool)
	tokenRepo := db.NewTokenRepo(pool)

	// Create auth handler.
	googleOAuth := auth.NewGoogleOAuthConfig(cfg)
	gitHubOAuth := auth.NewGitHubOAuthConfig(cfg)
	authHandler := auth.NewAuthHandler(googleOAuth, gitHubOAuth, cfg.JWTSecret, cfg.BaseURL, cfg.Env, userRepo, tokenRepo)

	// Create data handlers.
	taskHandler := handlers.NewTaskHandler(taskRepo)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)
	sessionHandler := handlers.NewSessionHandler(sessionRepo)
	noteHandler := handlers.NewNoteHandler(noteRepo)

	// Auth middleware.
	requireAuth := auth.RequireAuth(cfg.JWTSecret)

	mux := http.NewServeMux()

	// Auth routes (no auth middleware needed).
	auth.RegisterRoutes(mux, authHandler)

	// Protected API routes.
	mux.Handle("GET /api/tasks", requireAuth(http.HandlerFunc(taskHandler.List)))
	mux.Handle("POST /api/tasks", requireAuth(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("PUT /api/tasks/{id}", requireAuth(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /api/tasks/{id}", requireAuth(http.HandlerFunc(taskHandler.Delete)))

	mux.Handle("GET /api/settings", requireAuth(http.HandlerFunc(settingsHandler.Get)))
	mux.Handle("PUT /api/settings", requireAuth(http.HandlerFunc(settingsHandler.Put)))

	mux.Handle("GET /api/session", requireAuth(http.HandlerFunc(sessionHandler.Get)))
	mux.Handle("POST /api/session/start", requireAuth(http.HandlerFunc(sessionHandler.Start)))
	mux.Handle("POST /api/session/pause", requireAuth(http.HandlerFunc(sessionHandler.Pause)))
	mux.Handle("POST /api/session/stop", requireAuth(http.HandlerFunc(sessionHandler.Stop)))
	mux.Handle("PUT /api/session/totals", requireAuth(http.HandlerFunc(sessionHandler.UpdateTotals)))

	mux.Handle("GET /api/stats/completions", requireAuth(http.HandlerFunc(taskHandler.Completions)))
	mux.Handle("GET /api/stats/estimation", requireAuth(http.HandlerFunc(taskHandler.Estimation)))

	mux.Handle("GET /api/notes/{date}", requireAuth(http.HandlerFunc(noteHandler.Get)))
	mux.Handle("PUT /api/notes/{date}", requireAuth(http.HandlerFunc(noteHandler.Upsert)))

	// Static files.
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("sub static FS: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	// Middleware stack: SecurityHeaders → RateLimit → CORS → CSRF → routes
	secureCookie := cfg.Env == "production"
	csrfMiddleware := middleware.CSRFProtect(secureCookie)
	corsMiddleware := middleware.CORS([]string{cfg.BaseURL})
	rateLimitMiddleware := middleware.RateLimit(cfg.RateLimit, cfg.RateBurst)
	securityHeaders := middleware.SecurityHeaders()
	handler := middleware.Logging(securityHeaders(rateLimitMiddleware(corsMiddleware(csrfMiddleware(mux)))))

	addr := ":" + cfg.Port
	log.Printf("Web mode: listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// runDesktopMode starts the original JSON-backed desktop app with WebView2.
func runDesktopMode() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("os.Executable: %v", err)
	}
	dataDir = filepath.Join(filepath.Dir(exe), "data")
	stateFile = filepath.Join(dataDir, "state.json")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	loadState()

	mux := http.NewServeMux()
	registerDesktopRoutes(mux)

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
	ready := make(chan struct{})
	go func() {
		close(ready)
		log.Printf("Desktop mode: listening on %s", url)
		if err := http.Serve(ln, corsMiddleware(mux)); err != nil {
			log.Printf("server: %v", err)
		}
	}()
	<-ready

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

	hwnd := uintptr(w.Window())

	user32 := syscall.NewLazyDLL("user32.dll")
	setWindowPos := user32.NewProc("SetWindowPos")

	const (
		hwndTopmost    = ^uintptr(0)
		hwndNotTopmost = ^uintptr(0) - 1
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

	dwmapi := syscall.NewLazyDLL("dwmapi.dll")
	dwmSetAttr := dwmapi.NewProc("DwmSetWindowAttribute")

	type titleBarReq struct {
		IsDark bool   `json:"isDark"`
		Bg     string `json:"bg"`
	}
	w.Bind("goSetTitleBar", func(req titleBarReq) error {
		bg := hexToCOLORREF(req.Bg)
		dark := uint32(0)
		if req.IsDark {
			dark = 1
		}
		textColor := uint32(0x00DDDDDD)
		if !req.IsDark {
			textColor = 0x00222222
		}
		dwmSetAttr.Call(hwnd, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwmSetAttr.Call(hwnd, 35, uintptr(unsafe.Pointer(&bg)), 4)
		dwmSetAttr.Call(hwnd, 36, uintptr(unsafe.Pointer(&textColor)), 4)
		return nil
	})

	w.Bind("goNotify", func(req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}) error {
		script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime]|Out-Null
[Windows.Data.Xml.Dom.XmlDocument,Windows.Data.Xml.Dom.XmlDocument,ContentType=WindowsRuntime]|Out-Null
$t=[Windows.UI.Notifications.ToastTemplateType]::ToastImageAndText02
$x=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($t)
$x.GetElementsByTagName('text')[0].AppendChild($x.CreateTextNode('%s'))|Out-Null
$x.GetElementsByTagName('text')[1].AppendChild($x.CreateTextNode('%s'))|Out-Null
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Study Session').Show(
  [Windows.UI.Notifications.ToastNotification]::new($x))
`, req.Title, req.Message)
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
		cmd.Start()
		return nil
	})

	w.SetSize(393, 852, webview.HintNone)
	w.Navigate(url)
	w.Run()
}
