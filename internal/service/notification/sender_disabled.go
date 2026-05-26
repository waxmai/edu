package notification

import "context"

type DisabledSender struct{}

func (DisabledSender) Send(ctx context.Context, message Message) (*SendResult, error) {
	return nil, NewSendError(ErrorConfig, "notification delivery is disabled", nil)
}
