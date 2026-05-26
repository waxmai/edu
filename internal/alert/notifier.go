package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Event struct {
	Project    string
	Env        string
	Severity   string
	DedupKey   string
	Title      string
	Message    string
	TraceID    string
	Host       string
	URI        string
	Method     string
	Suggestion string
	OccurredAt time.Time
	Extra      map[string]any
}

type AlertNotifier interface {
	Notify(ctx context.Context, event Event) error
}

type noopNotifier struct{}

func (noopNotifier) Notify(context.Context, Event) error { return nil }

type stdoutNotifier struct{}

func (stdoutNotifier) Notify(ctx context.Context, event Event) error {
	body, _ := json.Marshal(eventPayload(event))
	fmt.Printf("ALERT %s\n", string(body))
	return nil
}

type webhookNotifier struct {
	kind    string
	url     string
	client  *http.Client
	timeout time.Duration
}

func newWebhookNotifier(kind, url string, timeout time.Duration) AlertNotifier {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &webhookNotifier{kind: kind, url: strings.TrimSpace(url), timeout: timeout, client: &http.Client{Timeout: timeout}}
}

func (n *webhookNotifier) Notify(ctx context.Context, event Event) error {
	if n == nil || n.url == "" {
		return nil
	}
	body, err := json.Marshal(n.webhookPayload(event))
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := n.client
	if client == nil {
		client = &http.Client{Timeout: n.timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *webhookNotifier) webhookPayload(event Event) map[string]any {
	text := renderEventText(event)
	switch strings.ToLower(n.kind) {
	case "feishu", "lark":
		return map[string]any{"msg_type": "text", "content": map[string]any{"text": text}}
	case "wecom", "wechat_work":
		return map[string]any{"msgtype": "text", "text": map[string]any{"content": text}}
	default:
		return map[string]any{"text": text, "event": eventPayload(event)}
	}
}

type cooldownNotifier struct {
	next     AlertNotifier
	cooldown time.Duration
	now      func() time.Time
	mu       sync.Mutex
	lastSent map[string]time.Time
}

func newCooldownNotifier(next AlertNotifier, cooldown time.Duration) AlertNotifier {
	if next == nil || cooldown <= 0 {
		return next
	}
	return &cooldownNotifier{next: next, cooldown: cooldown, now: time.Now, lastSent: map[string]time.Time{}}
}

func (n *cooldownNotifier) Notify(ctx context.Context, event Event) error {
	if n == nil || n.next == nil {
		return nil
	}
	key := strings.TrimSpace(event.DedupKey)
	if key == "" {
		key = event.Title + "|" + event.URI + "|" + event.Message
	}
	now := time.Now()
	if n.now != nil {
		now = n.now()
	}
	n.mu.Lock()
	last, ok := n.lastSent[key]
	if ok && now.Sub(last) < n.cooldown {
		n.mu.Unlock()
		return nil
	}
	n.lastSent[key] = now
	n.mu.Unlock()
	return n.next.Notify(ctx, event)
}

func NewNotifierFromEnv() AlertNotifier {
	channel := strings.ToLower(strings.TrimSpace(os.Getenv("ALERT_CHANNEL")))
	if channel == "" {
		channel = "stdout"
	}
	var base AlertNotifier
	switch channel {
	case "disabled", "none", "off":
		base = noopNotifier{}
	case "stdout":
		base = stdoutNotifier{}
	case "feishu", "lark":
		base = newWebhookNotifier("feishu", os.Getenv("ALERT_FEISHU_WEBHOOK"), envDurationSeconds("ALERT_WEBHOOK_TIMEOUT_SECONDS", 3))
	case "wecom", "wechat_work":
		base = newWebhookNotifier("wecom", os.Getenv("ALERT_WECOM_WEBHOOK"), envDurationSeconds("ALERT_WEBHOOK_TIMEOUT_SECONDS", 3))
	case "webhook":
		base = newWebhookNotifier("generic", os.Getenv("ALERT_WEBHOOK_URL"), envDurationSeconds("ALERT_WEBHOOK_TIMEOUT_SECONDS", 3))
	default:
		base = stdoutNotifier{}
	}
	return newCooldownNotifier(base, envDurationSeconds("ALERT_COOLDOWN_SECONDS", 300))
}

func renderEventText(event Event) string {
	severity := strings.ToUpper(firstNonEmpty(event.Severity, "warning"))
	title := firstNonEmpty(event.Title, "系统告警")
	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	lines := []string{
		fmt.Sprintf("[%s] %s", severity, title),
		fmt.Sprintf("项目：%s", firstNonEmpty(event.Project, "edu-schedule-system")),
		fmt.Sprintf("环境：%s", firstNonEmpty(event.Env, "unknown")),
		fmt.Sprintf("时间：%s", occurredAt.Format(time.RFC3339)),
	}
	if event.Message != "" {
		lines = append(lines, "内容："+event.Message)
	}
	if event.Method != "" || event.URI != "" {
		lines = append(lines, "请求："+strings.TrimSpace(event.Method+" "+event.URI))
	}
	if event.TraceID != "" {
		lines = append(lines, "Trace："+event.TraceID)
	}
	if event.Host != "" {
		lines = append(lines, "主机："+event.Host)
	}
	if event.Suggestion != "" {
		lines = append(lines, "建议："+event.Suggestion)
	}
	return strings.Join(lines, "\n")
}

func eventPayload(event Event) map[string]any {
	return map[string]any{
		"project":    event.Project,
		"env":        event.Env,
		"severity":   event.Severity,
		"dedupKey":   event.DedupKey,
		"title":      event.Title,
		"message":    event.Message,
		"traceId":    event.TraceID,
		"host":       event.Host,
		"uri":        event.URI,
		"method":     event.Method,
		"suggestion": event.Suggestion,
		"time":       event.OccurredAt,
		"extra":      event.Extra,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func envDurationSeconds(key string, def int) time.Duration {
	seconds := envInt(key, def)
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
