package auth

import (
	"context"
	"strings"

	"edu-schedule-system/configs"
	"go.uber.org/zap"
)

type smsStubRecoveryDeliverySender struct {
	provider string
	signName string
	logger   *zap.Logger
}

func newSMSStubRecoveryDeliverySender(cfg configs.Config, logger *zap.Logger) RecoveryDeliverySender {
	return &smsStubRecoveryDeliverySender{
		provider: strings.TrimSpace(cfg.Auth.RecoveryDelivery.SMS.Provider),
		signName: strings.TrimSpace(cfg.Auth.RecoveryDelivery.SMS.SignName),
		logger:   logger,
	}
}

func (s *smsStubRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	if s != nil && s.logger != nil {
		s.logger.Info("recovery-delivery-sms-stub",
			zap.String("challenge_id", message.ChallengeID),
			zap.Int32("user_id", message.UserID),
			zap.String("channel", message.Channel),
			zap.String("provider", s.provider),
			zap.String("sign_name", s.signName),
			zap.String("target", message.Target),
		)
	}
	return &RecoveryDeliveryResult{Accepted: true, Provider: fallbackProviderName(s.provider, "sms-stub"), TargetMasked: maskRecoveryTarget(message.Target, message.Channel)}, nil
}

func fallbackProviderName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
