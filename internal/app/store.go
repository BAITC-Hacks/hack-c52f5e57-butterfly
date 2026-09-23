// Package app implements private meeting review APIs, independent of model providers.
package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrUnauthorized = errors.New("создайте сессию перед работой с протоколами")
var ErrNotFound = errors.New("встреча не найдена")

const cookieName = "butterfly_session"
const maxUploadBytes = 50 << 20
const maxTranscriptChars = 300000

type record struct {
	Owner   string         `json:"owner"`
	Meeting map[string]any `json:"meeting"`
}
type state struct {
	Sessions map[string]bool   `json:"sessions"`
	Meetings map[string]record `json:"meetings"`
}

type App struct {
	mu         sync.Mutex
	mutationMu sync.Mutex
	dataDir    string
	state      state
	processing chan struct{}
	Export     func(map[string]any, string) ([]byte, error)
	OnCreated  func(map[string]any)
	OnReview   func(map[string]any, map[string]any)
}

func New(dataDir string) (*App, error) {
	if err := os.MkdirAll(filepath.Join(dataDir, "uploads"), 0700); err != nil {
		return nil, err
	}
	a := &App{dataDir: dataDir, processing: make(chan struct{}, 1), state: state{Sessions: map[string]bool{}, Meetings: map[string]record{}}}
	content, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if len(content) > 0 {
		if err = json.Unmarshal(content, &a.state); err != nil {
			return nil, fmt.Errorf("invalid stored state: %w", err)
		}
	}
	if a.state.Sessions == nil {
		a.state.Sessions = map[string]bool{}
	}
	if a.state.Meetings == nil {
		a.state.Meetings = map[string]record{}
	}
	recovered := false
	for id, row := range a.state.Meetings {
		if row.Meeting["status"] == "processing" {
			row.Meeting["status"] = "failed"
			row.Meeting["error"] = "Обработка прервана перезапуском сервера. Повторите попытку."
			addEvent(row.Meeting, "processing_failed", row.Meeting["error"].(string))
			a.state.Meetings[id] = row
			recovered = true
		}
	}
	if recovered {
		if err = a.persistLocked(); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func randomID() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure randomness unavailable")
	}
	return hex.EncodeToString(b[:])
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func clone(value map[string]any) map[string]any {
	raw, _ := json.Marshal(value)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}
func timestamp() string     { return time.Now().UTC().Format(time.RFC3339Nano) }
func text(value any) string { s, _ := value.(string); return s }
func entries(value any) []any {
	items, _ := value.([]any)
	if items == nil {
		return []any{}
	}
	return items
}
func object(value any) map[string]any { m, _ := value.(map[string]any); return m }
func publicMeeting(meeting map[string]any) map[string]any {
	out := clone(meeting)
	delete(out, "_audio_file")
	return out
}
func addEvent(meeting map[string]any, kind, message string) map[string]any {
	event := map[string]any{"id": randomID(), "type": kind, "message": message, "created_at": timestamp()}
	meeting["events"] = append(entries(meeting["events"]), event)
	return event
}

func (a *App) persistLocked() error {
	content, err := json.Marshal(a.state)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(a.dataDir, ".state-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(content)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(a.dataDir, "state.json"))
}
func (a *App) Owner(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" || len(cookie.Value) > 256 {
		return "", ErrUnauthorized
	}
	owner := hashToken(cookie.Value)
	a.mu.Lock()
	valid := a.state.Sessions[owner]
	a.mu.Unlock()
	if !valid {
		return "", ErrUnauthorized
	}
	return owner, nil
}
func (a *App) Get(owner, id string) (map[string]any, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	row, ok := a.state.Meetings[id]
	if !ok || row.Owner != owner {
		return nil, ErrNotFound
	}
	return clone(row.Meeting), nil
}
func (a *App) List(owner string) []map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []map[string]any{}
	for _, row := range a.state.Meetings {
		if row.Owner == owner {
			out = append(out, clone(row.Meeting))
		}
	}
	return out
}
func (a *App) Save(owner string, meeting map[string]any) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := text(meeting["id"])
	old, ok := a.state.Meetings[id]
	if !ok || old.Owner != owner {
		return ErrNotFound
	}
	a.state.Meetings[id] = record{owner, clone(meeting)}
	if err := a.persistLocked(); err != nil {
		a.state.Meetings[id] = old
		return err
	}
	return nil
}
func (a *App) create(owner string, meeting map[string]any) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := text(meeting["id"])
	if _, exists := a.state.Meetings[id]; exists {
		return errors.New("duplicate meeting")
	}
	a.state.Meetings[id] = record{owner, clone(meeting)}
	if err := a.persistLocked(); err != nil {
		delete(a.state.Meetings, id)
		return err
	}
	return nil
}
func (a *App) delete(owner, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	old, ok := a.state.Meetings[id]
	if !ok || old.Owner != owner {
		return ErrNotFound
	}
	delete(a.state.Meetings, id)
	if err := a.persistLocked(); err != nil {
		a.state.Meetings[id] = old
		return err
	}
	return nil
}
