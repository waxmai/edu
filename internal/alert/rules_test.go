package alert

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"edu-schedule-system/internal/proposal"
)

func testAlertMessage(message string, at string) *proposal.AlertMessage {
	return &proposal.AlertMessage{ProjectName: "edu", Env: "test", HOST: "localhost", URI: "/panic", Method: "GET", ErrorMessage: message, Time: at}
}

type captureNotifier struct{ events []Event }

func (n *captureNotifier) Notify(_ context.Context, event Event) error {
	n.events = append(n.events, event)
	return nil
}

func TestRuleEngineReadyCheckThresholdAndReset(t *testing.T) {
	base := &captureNotifier{}
	engine := NewRuleEngine(base, "edu", "test")
	engine.readyThreshold = 2
	engine.RecordReadyCheck(context.Background(), "mysql", errors.New("ping failed"))
	if len(base.events) != 0 {
		t.Fatalf("events = %d, want 0 before threshold", len(base.events))
	}
	engine.RecordReadyCheck(context.Background(), "mysql", errors.New("ping failed"))
	if len(base.events) != 1 {
		t.Fatalf("events = %d, want 1", len(base.events))
	}
	got := base.events[0]
	if got.Severity != "critical" || got.DedupKey != "ready-check:mysql" || !strings.Contains(got.Message, "mysql") {
		t.Fatalf("event = %#v", got)
	}
	engine.RecordReadyCheck(context.Background(), "mysql", nil)
	engine.RecordReadyCheck(context.Background(), "mysql", errors.New("ping failed"))
	if len(base.events) != 1 {
		t.Fatalf("events = %d, want no new event after reset first failure", len(base.events))
	}
}

func TestEventFromAlertMessageProducesReadablePanicEvent(t *testing.T) {
	msgTime := time.Now().Format("2006-01-02 15:04:05")
	event := eventFromAlertMessage(testAlertMessage("panic boom", msgTime))
	if event.Severity != "critical" || event.Title == "" || !strings.Contains(event.Message, "panic boom") || event.Suggestion == "" {
		t.Fatalf("event = %#v", event)
	}
}
