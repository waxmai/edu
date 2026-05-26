package alert

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"edu-schedule-system/internal/proposal"
)

type MetricsRuleMonitor struct {
	engine            *RuleEngine
	window            time.Duration
	minRequests       int
	fiveXXRatio       float64
	p95LatencySeconds float64
	now               func() time.Time
	mu                sync.Mutex
	requests          []metricsSample
}

type metricsSample struct {
	at          time.Time
	path        string
	method      string
	httpCode    int
	costSeconds float64
}

func NewMetricsRuleMonitor(engine *RuleEngine) *MetricsRuleMonitor {
	return &MetricsRuleMonitor{
		engine:            engine,
		window:            envDurationSeconds("ALERT_METRICS_WINDOW_SECONDS", 300),
		minRequests:       envInt("ALERT_METRICS_MIN_REQUESTS", 20),
		fiveXXRatio:       envFloat64("ALERT_HTTP_5XX_RATIO", 0.05),
		p95LatencySeconds: envFloat64("ALERT_HTTP_P95_SECONDS", 1.5),
		now:               time.Now,
	}
}

func (m *MetricsRuleMonitor) Record(ctx context.Context, msg *proposal.MetricsMessage) {
	if m == nil || msg == nil || m.engine == nil {
		return
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if m.window <= 0 {
		m.window = 5 * time.Minute
	}
	sample := metricsSample{at: now, path: msg.Path, method: msg.Method, httpCode: msg.HTTPCode, costSeconds: msg.CostSeconds}

	m.mu.Lock()
	cutoff := now.Add(-m.window)
	kept := m.requests[:0]
	for _, item := range m.requests {
		if item.at.After(cutoff) || item.at.Equal(cutoff) {
			kept = append(kept, item)
		}
	}
	kept = append(kept, sample)
	m.requests = kept
	snapshot := append([]metricsSample(nil), m.requests...)
	m.mu.Unlock()

	if len(snapshot) < m.minRequests {
		return
	}
	m.check5xx(ctx, snapshot)
	m.checkP95(ctx, snapshot)
}

func (m *MetricsRuleMonitor) check5xx(ctx context.Context, snapshot []metricsSample) {
	if m.fiveXXRatio <= 0 {
		return
	}
	count5xx := 0
	for _, item := range snapshot {
		if item.httpCode >= 500 && item.httpCode <= 599 {
			count5xx++
		}
	}
	ratio := float64(count5xx) / float64(len(snapshot))
	if ratio < m.fiveXXRatio {
		return
	}
	_ = m.engine.Notify(ctx, Event{
		Severity:   "critical",
		DedupKey:   "http:5xx-ratio",
		Title:      "HTTP 5xx 比例超过阈值",
		Message:    fmt.Sprintf("最近 %s 内 %d/%d 个请求为 5xx，比例 %.2f%%，阈值 %.2f%%", m.window, count5xx, len(snapshot), ratio*100, m.fiveXXRatio*100),
		Suggestion: "请检查应用错误日志、panic/recover 告警、数据库/缓存探活、上游依赖和最近发布。",
		Extra:      map[string]any{"windowSeconds": int(m.window.Seconds()), "requests": len(snapshot), "fiveXX": count5xx, "ratio": ratio},
	})
}

func (m *MetricsRuleMonitor) checkP95(ctx context.Context, snapshot []metricsSample) {
	if m.p95LatencySeconds <= 0 {
		return
	}
	values := make([]float64, 0, len(snapshot))
	for _, item := range snapshot {
		if item.costSeconds >= 0 {
			values = append(values, item.costSeconds)
		}
	}
	if len(values) == 0 {
		return
	}
	p95 := percentile(values, 0.95)
	if p95 < m.p95LatencySeconds {
		return
	}
	_ = m.engine.Notify(ctx, Event{
		Severity:   "warning",
		DedupKey:   "http:p95-latency",
		Title:      "p95 接口耗时超过阈值",
		Message:    fmt.Sprintf("最近 %s 内请求 p95 耗时 %.3fs，阈值 %.3fs", m.window, p95, m.p95LatencySeconds),
		Suggestion: "请检查慢 SQL、Redis 延迟、外部供应商耗时、实例负载和最近发布。",
		Extra:      map[string]any{"windowSeconds": int(m.window.Seconds()), "requests": len(values), "p95Seconds": p95, "thresholdSeconds": m.p95LatencySeconds},
	})
}

func percentile(values []float64, q float64) float64 {
	if len(values) == 0 {
		return 0
	}
	for i := 1; i < len(values); i++ {
		v := values[i]
		j := i - 1
		for ; j >= 0 && values[j] > v; j-- {
			values[j+1] = values[j]
		}
		values[j+1] = v
	}
	idx := int(q*float64(len(values)-1) + 0.999999)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(values) {
		idx = len(values) - 1
	}
	return values[idx]
}

func MetricsRecordHandlerWithAlerts(base func(*proposal.MetricsMessage), monitor *MetricsRuleMonitor) func(*proposal.MetricsMessage) {
	return func(msg *proposal.MetricsMessage) {
		if base != nil {
			base(msg)
		}
		if monitor != nil {
			monitor.Record(context.Background(), msg)
		}
	}
}

func MetricsMessage(method, path string, httpCode int, costSeconds float64) *proposal.MetricsMessage {
	return &proposal.MetricsMessage{Method: method, Path: path, HTTPCode: httpCode, CostSeconds: costSeconds, IsSuccess: httpCode >= http.StatusOK && httpCode < http.StatusBadRequest}
}
