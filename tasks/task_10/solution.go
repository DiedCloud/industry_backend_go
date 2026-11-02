package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task),
	}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	if strings.TrimSpace(title) == "" {
		return Task{}, ErrInvalidTitle
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := strconv.FormatUint(r.seq, 10)
	r.seq += 1
	task := Task{
		ID:        id,
		Title:     strings.TrimSpace(title),
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[id] = task
	return task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.tasks[id]
	return v, ok
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filteredTasks := make([]Task, 0, len(r.tasks))

	for _, t := range r.tasks {
		if t.Done == done {
			filteredTasks = append(filteredTasks, t)
		}
	}

	sort.Slice(filteredTasks, func(i, j int) bool {
		if filteredTasks[i].UpdatedAt.Equal(filteredTasks[j].UpdatedAt) {
			return filteredTasks[i].ID < filteredTasks[j].ID
		}

		return filteredTasks[i].UpdatedAt.After(filteredTasks[j].UpdatedAt)
	})

	return filteredTasks
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.tasks[id]
	if !ok {
		return v, ErrNotFound
	}

	if v.Done != done {
		v.Done = done
		v.UpdatedAt = r.clock.Now()
		r.tasks[id] = v
	}
	return v, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", h.handleCreate)
	mux.HandleFunc("GET /tasks/{id}", h.handleGet2)
	mux.HandleFunc("GET /tasks", h.handleList)
	mux.HandleFunc("PATCH /tasks/{id}", h.handlePatch2)
	mux.ServeHTTP(w, r)
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title *string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &request)
	if err != nil {
		writeJSON(w, 400, err.Error())
		return
	}
	if request.Title == nil {
		writeJSON(w, 400, ErrInvalidTitle)
		return
	}

	res, err := h.repo.Create(*request.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			writeJSON(w, 400, err.Error())
		} else {
			writeJSON(w, 500, "internal server error")
		}
		return
	}

	writeJSON(w, 201, res)
}
func (h *httpHandler) handleGet2(w http.ResponseWriter, r *http.Request) {
	h.handleGet(w, r, r.PathValue("id"))
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	resTask, ok := h.repo.Get(id)
	if !ok {
		writeJSON(w, 404, ErrNotFound)
		return
	}

	writeJSON(w, 200, resTask)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done := false
	values, exists := r.URL.Query()["done"]

	if exists {
		if len(values) != 1 {
			writeJSON(w, 400, "invalid done parameter")
			return
		}

		switch values[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			writeJSON(w, 400, "invalid done parameter")
			return
		}
	}

	resTask := h.repo.List(done)
	writeJSON(w, 200, resTask)
}
func (h *httpHandler) handlePatch2(w http.ResponseWriter, r *http.Request) {
	h.handlePatch(w, r, r.PathValue("id"))
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var request struct {
		Done *bool `json:"done"`
	}
	err := decodeStrictJSON(r.Body, &request)
	if err != nil {
		writeJSON(w, 400, err)
		return
	}
	if request.Done == nil {
		writeJSON(w, 400, "done is required")
		return
	}

	res, err := h.repo.SetDone(id, *request.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, 404, err.Error())
		} else {
			writeJSON(w, 500, "internal server error")
		}
		return
	}

	writeJSON(w, 200, res)
}
func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("body must contain exactly one JSON object")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
