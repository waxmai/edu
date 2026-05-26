package auth

import (
	"context"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/notification"
)

func buildRecoveryDeliverySender(mode string) RecoveryDeliverySender {
	return buildRecoveryDeliverySenderWithConfig(configs.Get(), mode)
}

func buildRecoveryDeliverySenderWithConfig(cfg configs.Config, mode string) RecoveryDeliverySender {
	notificationSender := notification.NewSender(notification.Config{
		Mode:           mode,
		TimeoutSeconds: cfg.Auth.RecoveryDelivery.TimeoutSeconds,
		MaxAttempts:    cfg.Auth.RecoveryDelivery.MaxAttempts,
		Email: notification.EmailConfig{
			Host:      cfg.Auth.RecoveryDelivery.Email.Host,
			Port:      cfg.Auth.RecoveryDelivery.Email.Port,
			Username:  cfg.Auth.RecoveryDelivery.Email.Username,
			Password:  cfg.Auth.RecoveryDelivery.Email.Password,
			From:      cfg.Auth.RecoveryDelivery.Email.From,
			LocalName: cfg.Auth.RecoveryDelivery.Email.LocalName,
		},
		SMS: notification.SMSConfig{
			Provider: cfg.Auth.RecoveryDelivery.SMS.Provider,
			SignName: cfg.Auth.RecoveryDelivery.SMS.SignName,
		},
	}, nil)
	return &notificationRecoveryDeliverySender{sender: notificationSender}
}

type disabledRecoveryDeliverySender struct{}

func (disabledRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	return nil, apperr.DependencyFailed("recovery delivery is disabled")
}

type notificationRecoveryDeliverySender struct {
	sender   notification.Sender
	recorder notification.Recorder
}

func (s *notificationRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	if s == nil || s.sender == nil {
		return nil, apperr.DependencyFailed("notification sender is unavailable")
	}
	notificationMessage := notification.Message{
		Channel:        message.Channel,
		Target:         message.Target,
		TemplateCode:   message.Template,
		Variables:      map[string]string{"code": message.Code, "username": message.Username},
		OrganizationID: message.OrganizationID,
		Scene:          notification.ScenePasswordRecovery,
		IdempotencyKey: message.ChallengeID,
	}
	result, err := notification.NewRecordingSender(s.sender, s.recorder, nil).Send(ctx, notificationMessage)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &RecoveryDeliveryResult{Accepted: result.Accepted, Provider: result.Provider, ExternalID: result.ExternalID, TargetMasked: result.TargetMasked}, nil
}
