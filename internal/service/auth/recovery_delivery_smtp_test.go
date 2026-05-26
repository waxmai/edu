package auth

import (
	"context"
	"errors"
	"testing"
)

type stubSMTPAttemptSender struct {
	results []*RecoveryDeliveryResult
	errs    []error
	calls   int
}

func (s *stubSMTPAttemptSender) next() (*RecoveryDeliveryResult, error) {
	idx := s.calls
	s.calls++
	if idx < len(s.errs) && s.errs[idx] != nil {
		return nil, s.errs[idx]
	}
	if idx < len(s.results) && s.results[idx] != nil {
		return s.results[idx], nil
	}
	return &RecoveryDeliveryResult{Accepted: true, Provider: "smtp", TargetMasked: "masked"}, nil
}

func TestSMTPRecoveryDeliveryRetryLogic(t *testing.T) {
	sender := &smtpRecoveryDeliverySender{maxAttempts: 2}
	attempt := &stubSMTPAttemptSender{errs: []error{timeoutErr{}, nil}, results: []*RecoveryDeliveryResult{nil, {Accepted: true, Provider: "smtp", TargetMasked: "masked"}}}
	result, err := sender.sendWithRetry(context.Background(), RecoveryDeliveryMessage{Target: "u@example.com"}, attempt.next)
	if err != nil {
		t.Fatalf("sendWithRetry() error = %v", err)
	}
	if result == nil || !result.Accepted {
		t.Fatalf("sendWithRetry() result = %#v, want accepted", result)
	}
	if attempt.calls != 2 {
		t.Fatalf("attempt calls = %d, want 2", attempt.calls)
	}
}

func TestSMTPRecoveryDeliveryDoesNotRetryConfigError(t *testing.T) {
	sender := &smtpRecoveryDeliverySender{maxAttempts: 2}
	attempt := &stubSMTPAttemptSender{errs: []error{errors.New("smtp auth failed")}}
	_, err := sender.sendWithRetry(context.Background(), RecoveryDeliveryMessage{Target: "u@example.com"}, attempt.next)
	if err == nil {
		t.Fatal("sendWithRetry() error = nil, want config failure")
	}
	if attempt.calls != 1 {
		t.Fatalf("attempt calls = %d, want 1", attempt.calls)
	}
}
