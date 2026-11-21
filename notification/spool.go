package notification

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Enqueue publishes a complete JSON file atomically. Identical pending events
// share a key, so a prolonged delivery failure does not duplicate them.
func Enqueue(dir string, event Event) error {
	key := event
	key.Time = key.Time.UTC()
	// Timestamp is excluded from identity; keep the first occurrence until delivery.
	rawKey, err := json.Marshal([]string{event.Kind, event.Record, event.OldIP, event.NewIP, event.Detail})
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%x.json", sha256.Sum256(rawKey))
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		return nil
	}
	raw, err := json.Marshal(key)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".event-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// Link provides atomic create-if-absent without overwriting another producer.
	if err := os.Link(tmp.Name(), filepath.Join(dir, name)); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

type QueuedEvent struct {
	Path  string
	Event Event
}

func ReadQueue(dir string) ([]QueuedEvent, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var queued []QueuedEvent
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var event Event
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, fmt.Errorf("evento inválido %s: %w", entry.Name(), err)
		}
		queued = append(queued, QueuedEvent{path, event})
	}
	sort.Slice(queued, func(i, j int) bool { return queued[i].Event.Time.Before(queued[j].Event.Time) })
	return queued, nil
}

type Event struct {
	Time   time.Time
	Kind   string
	Record string
	OldIP  string
	NewIP  string
	Detail string
}
