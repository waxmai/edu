package alert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebhookNotifierFormatsFeishuPayload(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode payload error = %v", err)
		}
	}))
	defer server.Close()

	notifier := newWebhookNotifier("feishu", server.URL, time.Second)
	err := notifier.Notify(context.Background(), Event{Project: "edu", Env: "test", Severity: "critical", Title: "测试告警", Message: "backup failed", URI: "/system/ready", Method: http.MethodGet, OccurredAt: time.Unix(1, 0).UTC()})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if got["msg_type"] != "text" {
		t.Fatalf("payload = %#v", got)
	}
	content, ok := got["content"].(map[string]any)
	if !ok || !strings.Contains(content["text"].(string), "测试告警") || !strings.Contains(content["text"].(string), "backup failed") {
		t.Fatalf("content = %#v", got["content"])
	}
}

func TestWebhookNotifierFormatsWeComPayload(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode payload error = %v", err)
		}
	}))
	defer server.Close()

	notifier := newWebhookNotifier("wecom", server.URL, time.Second)
	if err := notifier.Notify(context.Background(), Event{Title: "企微告警", Message: "mysql ping failed"}); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if got["msgtype"] != "text" {
		t.Fatalf("payload = %#v", got)
	}
	text, ok := got["text"].(map[string]any)
	if !ok || !strings.Contains(text["content"].(string), "企微告警") {
		t.Fatalf("text = %#v", got["text"])
	}
}

type countingNotifier struct{ count int }

func (n *countingNotifier) Notify(context.Context, Event) error {
	n.count++
	return nil
}

func TestCooldownNotifierDeduplicates(t *testing.T) {
	base := &countingNotifier{}
	cooldown := &cooldownNotifier{next: base, cooldown: time.Minute, now: func() time.Time { return time.Unix(100, 0) }, lastSent: map[string]time.Time{}}
	event := Event{DedupKey: "same-alert", Title: "same"}
	if err := cooldown.Notify(context.Background(), event); err != nil {
		t.Fatalf("Notify() first error = %v", err)
	}
	if err := cooldown.Notify(context.Background(), event); err != nil {
		t.Fatalf("Notify() second error = %v", err)
	}
	if base.count != 1 {
		t.Fatalf("count = %d, want 1", base.count)
	}
	cooldown.now = func() time.Time { return time.Unix(200, 0) }
	if err := cooldown.Notify(context.Background(), event); err != nil {
		t.Fatalf("Notify() after cooldown error = %v", err)
	}
	if base.count != 2 {
		t.Fatalf("count = %d, want 2", base.count)
	}
}
