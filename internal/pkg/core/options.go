package core

import (
	stdctx "context"

	"edu-schedule-system/internal/proposal"
)

type Option func(*option)

type ReadinessCheck func(stdctx.Context) error

type option struct {
	enablePProf      bool
	enableSwagger    bool
	enablePrometheus bool
	enableCors       bool
	alertNotify      proposal.AlertHandler
	alertRules       interface {
		RecordReadyCheck(stdctx.Context, string, error)
	}
	recordHandler   proposal.RecordHandler
	readinessChecks map[string]ReadinessCheck
}

// WithEnablePProf 启用 pprof
func WithEnablePProf() Option {
	return func(opt *option) {
		opt.enablePProf = true
	}
}

// WithEnableSwagger 启用 swagger
func WithEnableSwagger() Option {
	return func(opt *option) {
		opt.enableSwagger = true
	}
}

// WithEnablePrometheus 启用 prometheus
func WithEnablePrometheus(recordHandler proposal.RecordHandler) Option {
	return func(opt *option) {
		opt.enablePrometheus = true
		opt.recordHandler = recordHandler
	}
}

// WithAlertNotify 设置告警通知
func WithAlertNotify(alertHandler proposal.AlertHandler) Option {
	return func(opt *option) {
		opt.alertNotify = alertHandler
	}
}

func WithAlertRules(rules interface {
	RecordReadyCheck(stdctx.Context, string, error)
}) Option {
	return func(opt *option) {
		opt.alertRules = rules
	}
}

// WithEnableCors 设置支持跨域
func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
	}
}

func WithReadinessChecks(checks map[string]ReadinessCheck) Option {
	return func(opt *option) {
		opt.readinessChecks = checks
	}
}
