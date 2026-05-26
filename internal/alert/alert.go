package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"edu-schedule-system/internal/pkg/logpolicy"
	"edu-schedule-system/internal/pkg/timeutil"
	"edu-schedule-system/internal/proposal"

	"go.uber.org/zap"
)

type notifier struct {
	logger     *zap.Logger
	notifier   AlertNotifier
	thresholds thresholds
}

type thresholds struct {
	HTTPLatencySeconds     float64
	DBLatencySeconds       float64
	RedisLatencySeconds    float64
	ExternalLatencySeconds float64
	LoginFailures          int
}

func NotifyHandler(logger *zap.Logger) func(msg *proposal.AlertMessage) {
	n := &notifier{logger: logger, notifier: NewNotifierFromEnv(), thresholds: loadThresholds()}
	return n.handle
}

func loadThresholds() thresholds {
	return thresholds{
		HTTPLatencySeconds:     envFloat64("ALERT_HTTP_LATENCY_SECONDS", 1.5),
		DBLatencySeconds:       envFloat64("ALERT_DB_LATENCY_SECONDS", 0.3),
		RedisLatencySeconds:    envFloat64("ALERT_REDIS_LATENCY_SECONDS", 0.15),
		ExternalLatencySeconds: envFloat64("ALERT_EXTERNAL_LATENCY_SECONDS", 0.8),
		LoginFailures:          envInt("ALERT_LOGIN_FAILURES", 5),
	}
}

func (n *notifier) handle(msg *proposal.AlertMessage) {
	if msg == nil {
		return
	}
	event := eventFromAlertMessage(msg)
	payload := eventPayload(event)
	payload["error"] = msg.ErrorMessage
	payload["stack"] = msg.ErrorStack
	payload["time"] = msg.Time
	if n.logger != nil {
		n.logger.Error("alert.triggered", zap.Any("alert", payload))
	}
	if n.notifier != nil {
		ctx, cancel := context.WithTimeout(context.Background(), envDurationSeconds("ALERT_NOTIFY_TIMEOUT_SECONDS", 3))
		defer cancel()
		if err := n.notifier.Notify(ctx, event); err != nil && n.logger != nil {
			n.logger.Warn("alert.notify_failed", zap.Error(err))
		}
	}
	appendAlertRecord(payload)
}

func appendAlertRecord(payload map[string]any) {
	_ = os.MkdirAll("./logs", 0o755)
	_, _ = logpolicy.Apply(logpolicy.RetentionPolicy{
		LogDir:        "logs",
		Pattern:       "alerts.ndjson",
		MaxAgeDays:    30,
		MaxBytes:      4 * 1024 * 1024,
		RotationCount: 5,
	})
	f, err := os.OpenFile("./logs/alerts.ndjson", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	payload["recordedAt"] = time.Now().Format(time.RFC3339)
	body, _ := json.Marshal(payload)
	_, _ = f.Write(append(body, '\n'))
}

func eventFromAlertMessage(msg *proposal.AlertMessage) Event {
	event := Event{
		Project:    msg.ProjectName,
		Env:        msg.Env,
		Severity:   "critical",
		Title:      "服务 panic / recover 事件",
		Message:    fmt.Sprint(msg.ErrorMessage),
		TraceID:    msg.TraceID,
		Host:       msg.HOST,
		URI:        msg.URI,
		Method:     msg.Method,
		Suggestion: "请检查服务日志、错误堆栈、依赖探活和最近发布变更。",
		Extra: map[string]any{
			"errorStack": msg.ErrorStack,
			"time":       msg.Time,
		},
	}
	if msg.Time != "" {
		if parsed, err := time.Parse(timeutil.CSTLayout, msg.Time); err == nil {
			event.OccurredAt = parsed
		}
	}
	event.DedupKey = strings.TrimSpace(event.Method + " " + event.URI + " " + event.Message)
	return event
}

func envFloat64(key string, def float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return def
	}
	return parsed
}

func envInt(key string, def int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}
