package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"edu-schedule-system/internal/pkg/logpolicy"
)

type FailureAlertSink interface {
	NotifyNotificationFailure(record DeliveryRecord, threshold int, window time.Duration)
}

type FileFailureAlertSink struct {
	Path string
}

func NewFileFailureAlertSink(path string) FailureAlertSink {
	path = strings.TrimSpace(path)
	if path == "" {
		path = filepath.Join("logs", "notification-alerts.ndjson")
	}
	return FileFailureAlertSink{Path: path}
}

func (s FileFailureAlertSink) NotifyNotificationFailure(record DeliveryRecord, threshold int, window time.Duration) {
	path := s.Path
	if strings.TrimSpace(path) == "" {
		path = filepath.Join("logs", "notification-alerts.ndjson")
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_, _ = logpolicy.Apply(logpolicy.RetentionPolicy{LogDir: filepath.Dir(path), Pattern: filepath.Base(path), MaxAgeDays: 30, MaxBytes: 4 * 1024 * 1024, RotationCount: 5})
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	payload := map[string]any{
		"recordedAt":     time.Now().Format(time.RFC3339),
		"type":           "notification_failure_threshold",
		"threshold":      threshold,
		"windowSeconds":  int64(window.Seconds()),
		"organizationId": record.OrganizationID,
		"scene":          record.Scene,
		"channel":        record.Channel,
		"templateCode":   record.TemplateCode,
		"provider":       record.Provider,
		"targetMasked":   record.TargetMasked,
		"errorCode":      record.ErrorKind,
		"errorMessage":   record.ErrorMessage,
		"requestId":      firstNonEmpty(record.ExternalID, record.IdempotencyKey),
	}
	body, _ := json.Marshal(payload)
	_, _ = f.Write(append(body, '\n'))
}

type FailureAlertingRecorder struct {
	next      Recorder
	sink      FailureAlertSink
	window    time.Duration
	threshold int
	now       func() time.Time
	failures  []time.Time
}

func NewFailureAlertingRecorder(next Recorder, sink FailureAlertSink) Recorder {
	threshold := envInt("NOTIFICATION_FAILURE_ALERT_THRESHOLD", 5)
	windowSeconds := envInt("NOTIFICATION_FAILURE_ALERT_WINDOW_SECONDS", 300)
	if threshold <= 0 {
		return next
	}
	if windowSeconds <= 0 {
		windowSeconds = 300
	}
	if sink == nil {
		sink = NewFileFailureAlertSink(os.Getenv("NOTIFICATION_FAILURE_ALERT_PATH"))
	}
	return &FailureAlertingRecorder{next: next, sink: sink, threshold: threshold, window: time.Duration(windowSeconds) * time.Second, now: time.Now}
}

func (r *FailureAlertingRecorder) Record(ctx context.Context, record DeliveryRecord) error {
	if r == nil {
		return nil
	}
	if strings.EqualFold(record.Status, "error") {
		r.captureFailure(record)
	}
	if r.next != nil {
		return r.next.Record(ctx, record)
	}
	return nil
}

func (r *FailureAlertingRecorder) captureFailure(record DeliveryRecord) {
	now := time.Now()
	if r.now != nil {
		now = r.now()
	}
	cutoff := now.Add(-r.window)
	kept := r.failures[:0]
	for _, ts := range r.failures {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	r.failures = append(kept, now)
	if r.threshold > 0 && len(r.failures) >= r.threshold {
		if r.sink != nil {
			r.sink.NotifyNotificationFailure(record, r.threshold, r.window)
		}
		r.failures = r.failures[:0]
	}
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

func (r *FailureAlertingRecorder) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("FailureAlertingRecorder(threshold=%d, window=%s)", r.threshold, r.window)
}
