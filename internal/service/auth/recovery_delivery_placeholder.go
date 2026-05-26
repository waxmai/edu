package auth

import (
	"context"

	"go.uber.org/zap"
)

type auditOnlyRecoveryDeliverySender struct {
	logger *zap.Logger
}

func newAuditOnlyRecoveryDeliverySender(logger *zap.Logger) RecoveryDeliverySender {
	return &auditOnlyRecoveryDeliverySender{logger: logger}
}

func (s *auditOnlyRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	if s != nil && s.logger != nil {
		s.logger.Info("recovery-delivery-placeholder",
			zap.String("challenge_id", message.ChallengeID),
			zap.Int32("user_id", message.UserID),
			zap.String("channel", message.Channel),
			zap.String("target", message.Target),
			zap.String("template", message.Template),
		)
	}
	return &RecoveryDeliveryResult{Accepted: true, Provider: "placeholder-audit", TargetMasked: message.Target}, nil
}
