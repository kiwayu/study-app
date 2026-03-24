package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

var _ = sort.Strings // reserved for Task 4

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

// ---- Main -------------------------------------------------------------------

func main() {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	loadState()

	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("sub static FS: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
