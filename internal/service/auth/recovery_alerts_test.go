package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendRecoveryAlertRecordWritesNDJSON(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	appendRecoveryAlertRecord(map[string]any{"challengeId": "c1", "userId": 1, "channel": "email", "error": "smtp recovery delivery timed out"})
	body, err := os.ReadFile(filepath.Join("logs", "recovery-alerts.ndjson"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "challengeId") || !strings.Contains(text, "smtp recovery delivery timed out") {
		t.Fatalf("alert log content = %q", text)
	}
}

func TestAppendRecoveryDeliveryFailureAlertIncludesOperationalFields(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	challenge := &authRecoveryChallenge{ID: "challenge-1", UserID: 7, Username: "recoverme", Status: recoveryStatusExpired, ExpiresAt: time.Date(2026, 5, 16, 1, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	user := &sysUser{RoleCode: "platform_admin"}
	appendRecoveryDeliveryFailureAlert(challenge, user, "email", "recover@example.local", os.ErrDeadlineExceeded)

	body, err := os.ReadFile(filepath.Join("logs", "recovery-alerts.ndjson"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(body)
	for _, want := range []string{"challengeStatus", "roleCode", "recoverme", "re***@example.local", "expiresAt"} {
		if !strings.Contains(text, want) {
			t.Fatalf("alert log content missing %q: %q", want, text)
		}
	}
}
