package notification

import "strings"

func fallbackProviderName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func MaskTarget(target, channel string) string {
	target = strings.TrimSpace(target)
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case ChannelEmail:
		at := strings.Index(target, "@")
		if at <= 1 {
			return "***"
		}
		return target[:1] + "***" + target[at:]
	case ChannelSMS:
		if len(target) <= 4 {
			return "****"
		}
		return "****" + target[len(target)-4:]
	default:
		if target == "" {
			return ""
		}
		return "***"
	}
}
