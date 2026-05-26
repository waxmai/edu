package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender struct {
	host        string
	port        int
	username    string
	password    string
	from        string
	timeout     time.Duration
	startTLS    bool
	maxAttempts int
	localName   string
}

func NewSMTPSender(cfg EmailConfig, timeoutSeconds int64, maxAttempts int) Sender {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return &SMTPSender{host: cfg.Host, port: cfg.Port, username: cfg.Username, password: cfg.Password, from: cfg.From, timeout: time.Duration(timeoutSeconds) * time.Second, startTLS: true, maxAttempts: maxAttempts, localName: cfg.LocalName}
}

func (s *SMTPSender) Send(ctx context.Context, message Message) (*SendResult, error) {
	if strings.TrimSpace(message.Target) == "" {
		return nil, NewSendError(ErrorInvalidTarget, "notification target is required", nil)
	}
	if strings.TrimSpace(message.Variables["code"]) == "" {
		return nil, NewSendError(ErrorProviderRejected, "notification template variable code is required", nil)
	}
	return s.sendWithRetry(ctx, message, func() (*SendResult, error) { return s.sendOnce(ctx, message) })
}

func (s *SMTPSender) sendWithRetry(ctx context.Context, message Message, attempt func() (*SendResult, error)) (*SendResult, error) {
	attempts := s.maxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for i := 1; i <= attempts; i++ {
		result, err := attempt()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !shouldRetrySMTPError(classifySMTPError(err)) || i == attempts {
			break
		}
	}
	kind := classifySMTPError(lastErr)
	switch kind {
	case ErrorConfig:
		return nil, NewSendError(ErrorConfig, "smtp notification configuration failed", lastErr)
	case ErrorTimeout:
		return nil, NewSendError(ErrorTimeout, "smtp notification timed out", lastErr)
	case ErrorProviderRejected:
		return nil, NewSendError(ErrorProviderRejected, "smtp notification temporary failure", lastErr)
	default:
		return nil, NewSendError(ErrorUnknown, "smtp notification failed", lastErr)
	}
}

func (s *SMTPSender) sendOnce(ctx context.Context, message Message) (*SendResult, error) {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	dialer := &net.Dialer{Timeout: s.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(s.timeout))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return nil, err
	}
	defer client.Quit()
	if strings.TrimSpace(s.localName) != "" {
		_ = client.Hello(s.localName)
	}
	if s.startTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
				return nil, err
			}
		}
	}
	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
			return nil, err
		}
	}
	if err := client.Mail(s.from); err != nil {
		return nil, err
	}
	if err := client.Rcpt(message.Target); err != nil {
		return nil, err
	}
	writer, err := client.Data()
	if err != nil {
		return nil, err
	}

	code := message.Variables["code"]
	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("To: %s\r\n", message.Target))
	body.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	body.WriteString("Subject: Password recovery verification code\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	body.WriteString(fmt.Sprintf("Your verification code is: %s\n", code))
	body.WriteString("If you did not request a password reset, please ignore this email.\n")
	if _, err := writer.Write(body.Bytes()); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return &SendResult{Accepted: true, Provider: "smtp", TargetMasked: MaskTarget(message.Target, ChannelEmail)}, nil
}

func classifySMTPError(err error) string {
	if err == nil {
		return ErrorUnknown
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return ErrorTimeout
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "missing") || strings.Contains(text, "no such host") || strings.Contains(text, "unknown authority"):
		return ErrorConfig
	case strings.Contains(text, "auth") || strings.Contains(text, "username") || strings.Contains(text, "password"):
		return ErrorAuthFailed
	case strings.Contains(text, "timeout") || strings.Contains(text, "deadline"):
		return ErrorTimeout
	case strings.Contains(text, "rate") || strings.Contains(text, "too many"):
		return ErrorRateLimited
	case strings.Contains(text, "temporary") || strings.Contains(text, "try again"):
		return ErrorProviderRejected
	default:
		return ErrorUnknown
	}
}

func shouldRetrySMTPError(kind string) bool {
	return kind == ErrorTimeout || kind == ErrorRateLimited || kind == ErrorProviderRejected || kind == ErrorUnknown
}
