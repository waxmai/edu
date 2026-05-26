package auth

import (
	"context"

	"edu-schedule-system/internal/service/notification"
)

type RecoveryDeliveryMessage struct {
	ChallengeID    string
	UserID         int32
	Username       string
	OrganizationID int32
	Channel        string
	Target         string
	Code           string
	Template       string
}

type RecoveryDeliveryResult struct {
	Accepted     bool
	Provider     string
	ExternalID   string
	TargetMasked string
}

type RecoveryDeliverySender interface {
	Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error)
}

type recoveryDeliveryRecorder interface {
	Record(ctx context.Context, record notification.DeliveryRecord) error
}
