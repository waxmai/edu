package notification

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

type DeliveryRecord struct {
	Scene          string
	Channel        string
	TemplateCode   string
	OrganizationID int32
	Provider       string
	ExternalID     string
	TargetMasked   string
	IdempotencyKey string
	Status         string
	ErrorKind      string
	ErrorMessage   string
	RecordedAt     time.Time
}

type Recorder interface {
	Record(ctx context.Context, record DeliveryRecord) error
}

type RecorderFunc func(ctx context.Context, record DeliveryRecord) error

func (f RecorderFunc) Record(ctx context.Context, record DeliveryRecord) error {
	if f == nil {
		return nil
	}
	return f(ctx, record)
}

type RecordingSender struct {
	next     Sender
	recorder Recorder
	logger   *zap.Logger
}

func NewRecordingSender(next Sender, recorder Recorder, logger *zap.Logger) Sender {
	return &RecordingSender{next: next, recorder: recorder, logger: logger}
}

func (s *RecordingSender) Send(ctx context.Context, message Message) (*SendResult, error) {
	if s == nil || s.next == nil {
		err := NewSendError(ErrorConfig, "notification sender is unavailable", nil)
		s.record(ctx, message, nil, err)
		return nil, err
	}
	result, err := s.next.Send(ctx, message)
	if err != nil {
		s.record(ctx, message, result, err)
		return nil, err
	}
	if result == nil || !result.Accepted {
		err := NewSendError(ErrorProviderRejected, "notification delivery rejected", nil)
		s.record(ctx, message, result, err)
		return result, err
	}
	s.record(ctx, message, result, nil)
	return result, nil
}

func (s *RecordingSender) record(ctx context.Context, message Message, result *SendResult, sendErr error) {
	if s == nil {
		return
	}
	record := DeliveryRecord{
		Scene:          message.Scene,
		Channel:        message.Channel,
		TemplateCode:   message.TemplateCode,
		OrganizationID: message.OrganizationID,
		IdempotencyKey: message.IdempotencyKey,
		Status:         "accepted",
		RecordedAt:     time.Now(),
	}
	if result != nil {
		record.Provider = result.Provider
		record.ExternalID = result.ExternalID
		record.TargetMasked = result.TargetMasked
	}
	if record.TargetMasked == "" {
		record.TargetMasked = MaskTarget(message.Target, message.Channel)
	}
	if sendErr != nil {
		record.Status = "error"
		record.ErrorKind = ErrorKind(sendErr)
		record.ErrorMessage = sendErr.Error()
	}
	if s.recorder != nil {
		if err := s.recorder.Record(ctx, record); err != nil && s.logger != nil {
			s.logger.Warn("notification-record-failed", zap.Error(err), zap.String("scene", record.Scene), zap.String("channel", record.Channel), zap.String("status", record.Status))
		}
	}
	if s.logger != nil {
		fields := []zap.Field{
			zap.String("scene", record.Scene),
			zap.String("channel", record.Channel),
			zap.String("template", record.TemplateCode),
			zap.Int32("organization_id", record.OrganizationID),
			zap.String("provider", record.Provider),
			zap.String("target", record.TargetMasked),
			zap.String("status", record.Status),
			zap.String("error_kind", record.ErrorKind),
		}
		if record.Status == "error" {
			s.logger.Warn("notification-delivery", fields...)
		} else {
			s.logger.Info("notification-delivery", fields...)
		}
	}
}

func ErrorKind(err error) string {
	if err == nil {
		return ""
	}
	var sendErr *SendError
	if errors.As(err, &sendErr) && sendErr.Kind != "" {
		return sendErr.Kind
	}
	return ErrorUnknown
}
