package notification

import (
	"strings"

	"go.uber.org/zap"
)

func NewSender(cfg Config, logger *zap.Logger) Sender {
	return newBaseSender(cfg, logger)
}

func newBaseSender(cfg Config, logger *zap.Logger) Sender {
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "disabled":
		return DisabledSender{}
	case "email":
		return NewSMTPSender(cfg.Email, cfg.TimeoutSeconds, cfg.MaxAttempts)
	case "sms", "sms_stub", "stub_sms":
		return NewSMSStubSender(cfg.SMS, logger)
	case "placeholder":
		fallthrough
	default:
		return NewPlaceholderSender(logger)
	}
}
