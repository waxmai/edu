package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/service/apperr"
)

type smtpRecoveryDeliverySender struct {
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

func newSMTPRecoveryDeliverySender(cfg configs.Config) RecoveryDeliverySender {
	return &smtpRecoveryDeliverySender{
		host:        cfg.Auth.RecoveryDelivery.Email.Host,
		port:        cfg.Auth.RecoveryDelivery.Email.Port,
		username:    cfg.Auth.RecoveryDelivery.Email.Username,
		password:    cfg.Auth.RecoveryDelivery.Email.Password,
		from:        cfg.Auth.RecoveryDelivery.Email.From,
		timeout:     time.Duration(cfg.Auth.RecoveryDelivery.TimeoutSeconds) * time.Second,
		startTLS:    true,
		maxAttempts: cfg.Auth.RecoveryDelivery.MaxAttempts,
		localName:   cfg.Auth.RecoveryDelivery.Email.LocalName,
	}
}

func (s *smtpRecoveryDeliverySender) Send(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
	if strings.TrimSpace(message.Target) == "" {
		return nil, apperr.InvalidArgument("recovery delivery target is required")
	}
	if strings.TrimSpace(message.Code) == "" {
		return nil, apperr.InvalidArgument("recovery delivery code is required")
	}
	return s.sendWithRetry(ctx, message, func() (*RecoveryDeliveryResult, error) {
		return s.sendOnce(ctx, message)
	})
}

func (s *smtpRecoveryDeliverySender) sendWithRetry(ctx context.Context, message RecoveryDeliveryMessage, attempt func() (*RecoveryDeliveryResult, error)) (*RecoveryDeliveryResult, error) {
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
		if !shouldRetrySMTPDelivery(classifySMTPDeliveryError(err)) || i == attempts {
			break
		}
	}
	category := classifySMTPDeliveryError(lastErr)
	switch category {
	case recoveryDeliveryErrConfig:
		return nil, apperr.DependencyFailed("smtp recovery delivery configuration failed")
	case recoveryDeliveryErrTimeout:
		return nil, apperr.DependencyFailed("smtp recovery delivery timed out")
	case recoveryDeliveryErrTemporary:
		return nil, apperr.DependencyFailed("smtp recovery delivery temporary failure")
	default:
		return nil, apperr.DependencyFailed("smtp recovery delivery failed")
	}
}

func (s *smtpRecoveryDeliverySender) sendOnce(ctx context.Context, message RecoveryDeliveryMessage) (*RecoveryDeliveryResult, error) {
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

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("To: %s\r\n", message.Target))
	body.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	body.WriteString("Subject: Password recovery verification code\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	body.WriteString(fmt.Sprintf("Your verification code is: %s\n", message.Code))
	body.WriteString("If you did not request a password reset, please ignore this email.\n")

	if _, err := writer.Write(body.Bytes()); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return &RecoveryDeliveryResult{Accepted: true, Provider: "smtp", TargetMasked: message.Target}, nil
}
