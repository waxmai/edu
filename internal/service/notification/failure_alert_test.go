package notification

import (
	"testing"
	"time"
)

type captureFailureAlertSink struct {
	count     int
	last      DeliveryRecord
	threshold int
	window    time.Duration
}

func (s *captureFailureAlertSink) NotifyNotificationFailure(record DeliveryRecord, threshold int, window time.Duration) {
	s.count++
	s.last = record
	s.threshold = threshold
	s.window = window
}

func TestFailureAlertingRecorderTriggersAtThreshold(t *testing.T) {
	sink := &captureFailureAlertSink{}
	recorder := &FailureAlertingRecorder{sink: sink, threshold: 2, window: time.Minute, now: time.Now}
	if err := recorder.Record(nil, DeliveryRecord{Status: "error", Scene: ScenePasswordRecovery, Channel: ChannelEmail, ErrorKind: ErrorTimeout}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if sink.count != 0 {
		t.Fatalf("alert count after first error = %d, want 0", sink.count)
	}
	if err := recorder.Record(nil, DeliveryRecord{Status: "error", Scene: ScenePasswordRecovery, Channel: ChannelEmail, ErrorKind: ErrorTimeout}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if sink.count != 1 || sink.threshold != 2 || sink.last.ErrorKind != ErrorTimeout {
		t.Fatalf("sink = %#v", sink)
	}
}

func TestFailureAlertingRecorderIgnoresAccepted(t *testing.T) {
	sink := &captureFailureAlertSink{}
	recorder := &FailureAlertingRecorder{sink: sink, threshold: 1, window: time.Minute, now: time.Now}
	if err := recorder.Record(nil, DeliveryRecord{Status: "accepted", Scene: ScenePasswordRecovery, Channel: ChannelEmail}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if sink.count != 0 {
		t.Fatalf("alert count = %d, want 0", sink.count)
	}
}
