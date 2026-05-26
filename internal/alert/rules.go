package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RuleEngine struct {
	notifier       AlertNotifier
	project        string
	env            string
	mu             sync.Mutex
	readyFailures  map[string]int
	readyThreshold int
}

func NewRuleEngine(notifier AlertNotifier, project, env string) *RuleEngine {
	if notifier == nil {
		notifier = noopNotifier{}
	}
	return &RuleEngine{notifier: notifier, project: project, env: env, readyFailures: map[string]int{}, readyThreshold: envInt("ALERT_READY_FAILURES", 3)}
}

func NewRuleEngineFromEnv(project, env string) *RuleEngine {
	return NewRuleEngine(NewNotifierFromEnv(), project, env)
}

func (e *RuleEngine) Notify(ctx context.Context, event Event) error {
	if e == nil || e.notifier == nil {
		return nil
	}
	if event.Project == "" {
		event.Project = e.project
	}
	if event.Env == "" {
		event.Env = e.env
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	return e.notifier.Notify(ctx, event)
}

func (e *RuleEngine) RecordReadyCheck(ctx context.Context, name string, err error) {
	if e == nil || err == nil {
		if e != nil {
			e.mu.Lock()
			delete(e.readyFailures, name)
			e.mu.Unlock()
		}
		return
	}
	e.mu.Lock()
	e.readyFailures[name]++
	count := e.readyFailures[name]
	threshold := e.readyThreshold
	if threshold <= 0 {
		threshold = 3
	}
	e.mu.Unlock()
	if count < threshold {
		return
	}
	_ = e.Notify(ctx, Event{
		Severity:   "critical",
		DedupKey:   "ready-check:" + name,
		Title:      "服务就绪检查连续失败",
		Message:    fmt.Sprintf("依赖 %s 连续失败 %d 次：%v", name, count, err),
		URI:        "/system/ready",
		Method:     "GET",
		Suggestion: "请检查该依赖连接、凭据、网络和最近发布变更。",
	})
}

func (e *RuleEngine) NotifyPanic(ctx context.Context, event Event) {
	if event.Severity == "" {
		event.Severity = "critical"
	}
	if event.Title == "" {
		event.Title = "服务 panic / recover 事件"
	}
	if event.DedupKey == "" {
		event.DedupKey = "panic:" + event.Method + ":" + event.URI
	}
	if event.Suggestion == "" {
		event.Suggestion = "请立即查看错误堆栈、请求参数和最近发布变更。"
	}
	_ = e.Notify(ctx, event)
}

func (e *RuleEngine) NotifyBackupFailure(ctx context.Context, message string, extra map[string]any) {
	_ = e.Notify(ctx, Event{Severity: "critical", DedupKey: "backup:failure", Title: "数据库备份失败", Message: message, Suggestion: "请检查备份脚本、MySQL 连接、磁盘空间、对象存储配置和最近一次成功备份。", Extra: extra})
}

func (e *RuleEngine) NotifyDiskSpaceLow(ctx context.Context, message string, extra map[string]any) {
	_ = e.Notify(ctx, Event{Severity: "critical", DedupKey: "disk:space-low", Title: "磁盘空间不足", Message: message, Suggestion: "请清理日志/备份归档，确认磁盘扩容和日志轮转策略。", Extra: extra})
}

func (e *RuleEngine) NotifyNotificationFailureSpike(ctx context.Context, message string, extra map[string]any) {
	_ = e.Notify(ctx, Event{Severity: "warning", DedupKey: "notification:failure-spike", Title: "通知发送失败异常升高", Message: message, Suggestion: "请检查短信/邮件供应商状态、凭据、限流和模板审核状态。", Extra: extra})
}

func (e *RuleEngine) NotifyRecoveryFailureSpike(ctx context.Context, message string, extra map[string]any) {
	_ = e.Notify(ctx, Event{Severity: "warning", DedupKey: "recovery:delivery-failure-spike", Title: "密码恢复发送失败异常升高", Message: message, Suggestion: "请检查密码恢复通道配置、供应商限流、模板和目标地址有效性。", Extra: extra})
}
