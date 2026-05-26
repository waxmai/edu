package auth

import (
	"errors"
	"net"
	"strings"
)

type recoveryDeliveryErrorCategory string

const (
	recoveryDeliveryErrConfig    recoveryDeliveryErrorCategory = "config"
	recoveryDeliveryErrTimeout   recoveryDeliveryErrorCategory = "timeout"
	recoveryDeliveryErrTemporary recoveryDeliveryErrorCategory = "temporary"
	recoveryDeliveryErrPermanent recoveryDeliveryErrorCategory = "permanent"
)

func classifySMTPDeliveryError(err error) recoveryDeliveryErrorCategory {
	if err == nil {
		return recoveryDeliveryErrPermanent
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return recoveryDeliveryErrTimeout
		}
		if netErr.Temporary() {
			return recoveryDeliveryErrTemporary
		}
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(msg, "auth"), strings.Contains(msg, "username"), strings.Contains(msg, "password"), strings.Contains(msg, "from"):
		return recoveryDeliveryErrConfig
	case strings.Contains(msg, "timeout"):
		return recoveryDeliveryErrTimeout
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "broken pipe"), strings.Contains(msg, "reset by peer"), strings.Contains(msg, "temporar"):
		return recoveryDeliveryErrTemporary
	default:
		return recoveryDeliveryErrPermanent
	}
}

func shouldRetrySMTPDelivery(category recoveryDeliveryErrorCategory) bool {
	return category == recoveryDeliveryErrTimeout || category == recoveryDeliveryErrTemporary
}
