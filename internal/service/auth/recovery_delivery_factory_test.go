package auth

import (
	"testing"

	"edu-schedule-system/configs"
)

func TestBuildRecoveryDeliverySenderReturnsNotificationAdapterForEmailMode(t *testing.T) {
	cfg := configs.Config{}
	cfg.Auth.RecoveryDelivery.Mode = "email"
	cfg.Auth.RecoveryDelivery.TimeoutSeconds = 10
	cfg.Auth.RecoveryDelivery.MaxAttempts = 2
	cfg.Auth.RecoveryDelivery.Email.Host = "smtp.example.com"
	cfg.Auth.RecoveryDelivery.Email.Port = 587
	cfg.Auth.RecoveryDelivery.Email.Username = "mailer"
	cfg.Auth.RecoveryDelivery.Email.Password = "secret"
	cfg.Auth.RecoveryDelivery.Email.From = "noreply@example.com"

	sender := buildRecoveryDeliverySenderWithConfig(cfg, cfg.Auth.RecoveryDelivery.Mode)
	adapter, ok := sender.(*notificationRecoveryDeliverySender)
	if !ok {
		t.Fatalf("buildRecoveryDeliverySenderWithConfig() = %T, want notificationRecoveryDeliverySender", sender)
	}
	if adapter.sender == nil {
		t.Fatalf("notification adapter sender is nil")
	}
}

func TestBuildRecoveryDeliverySenderReturnsNotificationAdapterForSMSMode(t *testing.T) {
	cfg := configs.Config{}
	cfg.Auth.RecoveryDelivery.Mode = "sms"
	cfg.Auth.RecoveryDelivery.SMS.Provider = "stub"
	cfg.Auth.RecoveryDelivery.SMS.SignName = "EduSchedule"

	sender := buildRecoveryDeliverySenderWithConfig(cfg, cfg.Auth.RecoveryDelivery.Mode)
	adapter, ok := sender.(*notificationRecoveryDeliverySender)
	if !ok {
		t.Fatalf("buildRecoveryDeliverySenderWithConfig() = %T, want notificationRecoveryDeliverySender", sender)
	}
	if adapter.sender == nil {
		t.Fatalf("notification adapter sender is nil")
	}
}
