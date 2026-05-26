package notification

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

type SMSStubSender struct {
	provider string
	signName string
	logger   *zap.Logger
}

func NewSMSStubSender(cfg SMSConfig, logger *zap.Logger) Sender {
	return &SMSStubSender{provider: strings.TrimSpace(cfg.Provider), signName: strings.TrimSpace(cfg.SignName), logger: logger}
}

func (s *SMSStubSender) Send(ctx context.Context, message Message) (*SendResult, error) {
	if strings.TrimSpace(message.Target) == "" {
		return nil, NewSendError(ErrorInvalidTarget, "notification target is required", nil)
	}
	if s != nil && s.logger != nil {
		s.logger.Info("notification-sms-stub",
			zap.String("scene", message.Scene),
			zap.String("channel", message.Channel),
			zap.String("provider", s.provider),
			zap.String("sign_name", s.signName),
			zap.String("target", MaskTarget(message.Target, message.Channel)),
			zap.String("template", message.TemplateCode),
		)
	}
	provider := "sms-stub"
	if s != nil {
		provider = fallbackProviderName(s.provider, provider)
	}
	return &SendResult{Accepted: true, Provider: provider, TargetMasked: MaskTarget(message.Target, ChannelSMS)}, nil
}
