package notification

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestQueueAtomicDedupAndRestart(t *testing.T) {
	dir := t.TempDir()
	event := Event{Time: time.Now(), Kind: "Inicio", Detail: "Servidor activo"}
	if err := Enqueue(dir, event); err != nil {
		t.Fatal(err)
	}
	event.Time = time.Now().Add(time.Second)
	if err := Enqueue(dir, event); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, ".event-incomplete"), []byte("{"), 0600)
	queued, err := ReadQueue(dir)
	if err != nil || len(queued) != 1 {
		t.Fatal(queued, err)
	}
	if err := os.Remove(queued[0].Path); err != nil {
		t.Fatal(err)
	}
	queued, err = ReadQueue(dir)
	if err != nil || len(queued) != 0 {
		t.Fatal(queued, err)
	}
}
