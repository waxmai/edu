package notification

import (
	"context"

	"go.uber.org/zap"
)

type PlaceholderSender struct {
	logger *zap.Logger
}

func NewPlaceholderSender(logger *zap.Logger) Sender {
	return &PlaceholderSender{logger: logger}
}

func (s *PlaceholderSender) Send(ctx context.Context, message Message) (*SendResult, error) {
	if s != nil && s.logger != nil {
		s.logger.Info("notification-placeholder",
			zap.String("scene", message.Scene),
			zap.String("channel", message.Channel),
			zap.String("template", message.TemplateCode),
			zap.Int32("organization_id", message.OrganizationID),
			zap.String("target", MaskTarget(message.Target, message.Channel)),
			zap.String("idempotency_key", message.IdempotencyKey),
		)
	}
	return &SendResult{Accepted: true, Provider: "placeholder-audit", TargetMasked: MaskTarget(message.Target, message.Channel)}, nil
}
