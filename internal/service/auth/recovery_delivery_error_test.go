package auth

import (
	"errors"
	"net"
	"testing"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestClassifySMTPDeliveryError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want recoveryDeliveryErrorCategory
	}{
		{name: "timeout", err: timeoutErr{}, want: recoveryDeliveryErrTimeout},
		{name: "config auth", err: errors.New("smtp auth failed"), want: recoveryDeliveryErrConfig},
		{name: "temporary conn refused", err: errors.New("connection refused"), want: recoveryDeliveryErrTemporary},
		{name: "permanent", err: errors.New("mailbox unavailable"), want: recoveryDeliveryErrPermanent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifySMTPDeliveryError(tc.err)
			if got != tc.want {
				t.Fatalf("classifySMTPDeliveryError() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestShouldRetrySMTPDelivery(t *testing.T) {
	if !shouldRetrySMTPDelivery(recoveryDeliveryErrTimeout) {
		t.Fatal("timeout should be retryable")
	}
	if !shouldRetrySMTPDelivery(recoveryDeliveryErrTemporary) {
		t.Fatal("temporary should be retryable")
	}
	if shouldRetrySMTPDelivery(recoveryDeliveryErrConfig) {
		t.Fatal("config should not be retryable")
	}
	if shouldRetrySMTPDelivery(recoveryDeliveryErrPermanent) {
		t.Fatal("permanent should not be retryable")
	}
}

var _ net.Error = timeoutErr{}
