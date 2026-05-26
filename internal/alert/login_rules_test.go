package alert

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLoginFailureMonitorAlertsAtThreshold(t *testing.T) {
	base := &captureNotifier{}
	engine := NewRuleEngine(base, "edu", "test")
	monitor := NewLoginFailureMonitor(engine)
	monitor.window = time.Minute
	monitor.threshold = 3
	monitor.now = func() time.Time { return time.Unix(100, 0) }

	monitor.Record(context.Background(), LoginFailureEvent{Username: "alice", ClientIP: "1.1.1.1", Reason: "bad_password"})
	monitor.Record(context.Background(), LoginFailureEvent{Username: "bob", ClientIP: "1.1.1.1", Reason: "bad_password"})
	if len(base.events) != 0 {
		t.Fatalf("events = %d, want 0 before threshold", len(base.events))
	}
	monitor.Record(context.Background(), LoginFailureEvent{Username: "carol", ClientIP: "2.2.2.2", Reason: "user_not_found"})
	if len(base.events) != 1 {
		t.Fatalf("events = %d, want 1", len(base.events))
	}
	if base.events[0].DedupKey != "auth:login-failure-spike" || !strings.Contains(base.events[0].Message, "登录失败") {
		t.Fatalf("event = %#v", base.events[0])
	}
}
