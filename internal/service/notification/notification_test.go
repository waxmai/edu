package notification

import (
	"context"
	"errors"
	"testing"
)

func TestPlaceholderSenderMasksTarget(t *testing.T) {
	sender := NewPlaceholderSender(nil)
	result, err := sender.Send(context.Background(), Message{Channel: ChannelEmail, Target: "user@example.com", TemplateCode: "password_recovery", Scene: ScenePasswordRecovery})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result == nil || !result.Accepted {
		t.Fatalf("Send() result = %#v, want accepted", result)
	}
	if result.TargetMasked == "user@example.com" || result.TargetMasked == "" {
		t.Fatalf("TargetMasked = %q, want masked", result.TargetMasked)
	}
}

func TestSMSStubRejectsEmptyTarget(t *testing.T) {
	sender := NewSMSStubSender(SMSConfig{Provider: "stub"}, nil)
	_, err := sender.Send(context.Background(), Message{Channel: ChannelSMS})
	var sendErr *SendError
	if !errors.As(err, &sendErr) {
		t.Fatalf("Send() error = %T, want *SendError", err)
	}
	if sendErr.Kind != ErrorInvalidTarget {
		t.Fatalf("SendError.Kind = %q, want %q", sendErr.Kind, ErrorInvalidTarget)
	}
}

func TestFactoryReturnsDisabledSender(t *testing.T) {
	sender := newBaseSender(Config{Mode: "disabled"}, nil)
	if _, ok := sender.(DisabledSender); !ok {
		t.Fatalf("newBaseSender(disabled) = %T, want DisabledSender", sender)
	}
}

func TestRecordingSenderRecordsSuccessAndFailure(t *testing.T) {
	var records []DeliveryRecord
	recorder := RecorderFunc(func(ctx context.Context, record DeliveryRecord) error {
		records = append(records, record)
		return nil
	})
	successSender := SenderFunc(func(ctx context.Context, message Message) (*SendResult, error) {
		return &SendResult{Accepted: true, Provider: "stub", TargetMasked: MaskTarget(message.Target, message.Channel)}, nil
	})
	sender := NewRecordingSender(successSender, recorder, nil)
	_, err := sender.Send(context.Background(), Message{Channel: ChannelEmail, Target: "user@example.com", TemplateCode: "password_recovery", Scene: ScenePasswordRecovery, IdempotencyKey: "k1"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	failureSender := SenderFunc(func(ctx context.Context, message Message) (*SendResult, error) {
		return nil, NewSendError(ErrorTimeout, "timeout", nil)
	})
	sender = NewRecordingSender(failureSender, recorder, nil)
	_, _ = sender.Send(context.Background(), Message{Channel: ChannelSMS, Target: "10000013800", TemplateCode: "password_recovery", Scene: ScenePasswordRecovery, IdempotencyKey: "k2"})
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	if records[0].Status != "accepted" || records[0].Provider != "stub" {
		t.Fatalf("success record = %#v", records[0])
	}
	if records[1].Status != "error" || records[1].ErrorKind != ErrorTimeout {
		t.Fatalf("failure record = %#v", records[1])
	}
}

type SenderFunc func(ctx context.Context, message Message) (*SendResult, error)

func (f SenderFunc) Send(ctx context.Context, message Message) (*SendResult, error) {
	return f(ctx, message)
}
