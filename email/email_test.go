package email

import (
	"orgmserver/notification"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSharedModeNeverSendsSMTPAndRetainsLocalFailures(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "queue")
	t.Setenv("NOTIFICATION_DIR", dir)
	svc := NewEmailService("Test", "invalid.smtp", 0, "from@example.test", "password", "to@example.test", false)
	if err := svc.SendStartupEmail("192.0.2.1"); err == nil {
		t.Fatal("missing queue should fail")
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := svc.RetryPending(); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendIPChangeEmail("192.0.2.2", "192.0.2.1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendReconnectionEmail("192.0.2.2", time.Minute); err != nil {
		t.Fatal(err)
	}
	queued, err := notification.ReadQueue(dir)
	if err != nil || len(queued) != 3 {
		t.Fatal(queued, err)
	}
	if len(svc.pending) != 0 {
		t.Fatal("queue not drained")
	}
}
