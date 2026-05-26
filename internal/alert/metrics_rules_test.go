package alert

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMetricsRuleMonitorAlertsOn5xxRatio(t *testing.T) {
	base := &captureNotifier{}
	engine := NewRuleEngine(base, "edu", "test")
	monitor := NewMetricsRuleMonitor(engine)
	monitor.window = time.Minute
	monitor.minRequests = 4
	monitor.fiveXXRatio = 0.5
	monitor.p95LatencySeconds = 0
	monitor.now = func() time.Time { return time.Unix(100, 0) }

	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 200, 0.1))
	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 500, 0.1))
	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 502, 0.1))
	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 200, 0.1))

	if len(base.events) != 1 {
		t.Fatalf("events = %d, want 1", len(base.events))
	}
	if base.events[0].DedupKey != "http:5xx-ratio" || !strings.Contains(base.events[0].Message, "5xx") {
		t.Fatalf("event = %#v", base.events[0])
	}
}

func TestMetricsRuleMonitorAlertsOnP95Latency(t *testing.T) {
	base := &captureNotifier{}
	engine := NewRuleEngine(base, "edu", "test")
	monitor := NewMetricsRuleMonitor(engine)
	monitor.window = time.Minute
	monitor.minRequests = 4
	monitor.fiveXXRatio = 0
	monitor.p95LatencySeconds = 1.0
	monitor.now = func() time.Time { return time.Unix(100, 0) }

	monitor.Record(context.Background(), MetricsMessage("GET", "/slow", 200, 0.1))
	monitor.Record(context.Background(), MetricsMessage("GET", "/slow", 200, 0.2))
	monitor.Record(context.Background(), MetricsMessage("GET", "/slow", 200, 1.5))
	monitor.Record(context.Background(), MetricsMessage("GET", "/slow", 200, 2.0))

	if len(base.events) != 1 {
		t.Fatalf("events = %d, want 1", len(base.events))
	}
	if base.events[0].DedupKey != "http:p95-latency" || !strings.Contains(base.events[0].Message, "p95") {
		t.Fatalf("event = %#v", base.events[0])
	}
}

func TestMetricsRuleMonitorDropsOldSamples(t *testing.T) {
	base := &captureNotifier{}
	engine := NewRuleEngine(base, "edu", "test")
	monitor := NewMetricsRuleMonitor(engine)
	monitor.window = time.Second
	monitor.minRequests = 2
	monitor.fiveXXRatio = 0.5
	monitor.p95LatencySeconds = 0
	now := time.Unix(100, 0)
	monitor.now = func() time.Time { return now }
	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 500, 0.1))
	now = now.Add(2 * time.Second)
	monitor.Record(context.Background(), MetricsMessage("GET", "/a", 200, 0.1))
	if len(base.events) != 0 {
		t.Fatalf("events = %d, want old sample dropped", len(base.events))
	}
}
