package notification

import "context"

const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"

	ScenePasswordRecovery = "password_recovery"

	ErrorTimeout          = "timeout"
	ErrorAuthFailed       = "auth_failed"
	ErrorProviderRejected = "provider_rejected"
	ErrorRateLimited      = "rate_limited"
	ErrorInvalidTarget    = "invalid_target"
	ErrorConfig           = "config"
	ErrorUnknown          = "unknown"
)

type Message struct {
	Channel        string            `json:"channel"`
	Target         string            `json:"target"`
	TemplateCode   string            `json:"templateCode"`
	Variables      map[string]string `json:"variables"`
	OrganizationID int32             `json:"organizationId"`
	Scene          string            `json:"scene"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

type SendResult struct {
	Accepted     bool   `json:"accepted"`
	Provider     string `json:"provider"`
	ExternalID   string `json:"externalId,omitempty"`
	TargetMasked string `json:"targetMasked"`
}

type SendError struct {
	Kind    string
	Message string
	Err     error
}

func (e *SendError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Kind
}

func (e *SendError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewSendError(kind, message string, err error) *SendError {
	return &SendError{Kind: kind, Message: message, Err: err}
}

type Sender interface {
	Send(ctx context.Context, message Message) (*SendResult, error)
}
