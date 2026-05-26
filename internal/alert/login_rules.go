package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LoginFailureEvent struct {
	Username  string
	ClientIP  string
	UserAgent string
	Reason    string
}

type LoginFailureMonitor struct {
	engine    *RuleEngine
	window    time.Duration
	threshold int
	now       func() time.Time
	mu        sync.Mutex
	failures  []loginFailureSample
}

type loginFailureSample struct {
	at       time.Time
	username string
	clientIP string
	reason   string
}

func NewLoginFailureMonitor(engine *RuleEngine) *LoginFailureMonitor {
	return &LoginFailureMonitor{
		engine:    engine,
		window:    envDurationSeconds("ALERT_LOGIN_FAILURE_WINDOW_SECONDS", 300),
		threshold: envInt("ALERT_LOGIN_FAILURES", 5),
		now:       time.Now,
	}
}

func (m *LoginFailureMonitor) Record(ctx context.Context, event LoginFailureEvent) {
	if m == nil || m.engine == nil {
		return
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if m.window <= 0 {
		m.window = 5 * time.Minute
	}
	threshold := m.threshold
	if threshold <= 0 {
		threshold = 5
	}

	m.mu.Lock()
	cutoff := now.Add(-m.window)
	kept := m.failures[:0]
	for _, item := range m.failures {
		if item.at.After(cutoff) || item.at.Equal(cutoff) {
			kept = append(kept, item)
		}
	}
	kept = append(kept, loginFailureSample{at: now, username: event.Username, clientIP: event.ClientIP, reason: event.Reason})
	m.failures = kept
	snapshot := append([]loginFailureSample(nil), m.failures...)
	m.mu.Unlock()

	if len(snapshot) < threshold {
		return
	}
	byReason := map[string]int{}
	byIP := map[string]int{}
	for _, item := range snapshot {
		byReason[item.reason]++
		if item.clientIP != "" {
			byIP[item.clientIP]++
		}
	}
	_ = m.engine.Notify(ctx, Event{
		Severity:   "warning",
		DedupKey:   "auth:login-failure-spike",
		Title:      "登录失败次数异常升高",
		Message:    fmt.Sprintf("最近 %s 内登录失败 %d 次，阈值 %d 次", m.window, len(snapshot), threshold),
		Suggestion: "请检查是否存在撞库/暴力破解、异常 IP、账号锁定策略和认证服务日志。",
		Extra:      map[string]any{"windowSeconds": int(m.window.Seconds()), "failures": len(snapshot), "threshold": threshold, "byReason": byReason, "byIP": byIP},
	})
}
