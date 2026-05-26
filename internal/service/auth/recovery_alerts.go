package auth

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"edu-schedule-system/internal/pkg/logpolicy"
)

func appendRecoveryAlertRecord(payload map[string]any) {
	_ = os.MkdirAll("./logs", 0o755)
	_, _ = logpolicy.Apply(logpolicy.RetentionPolicy{
		LogDir:        "logs",
		Pattern:       "recovery-alerts.ndjson",
		MaxAgeDays:    30,
		MaxBytes:      4 * 1024 * 1024,
		RotationCount: 5,
	})
	f, err := os.OpenFile(filepath.Join("logs", "recovery-alerts.ndjson"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	payload["recordedAt"] = time.Now().Format(time.RFC3339)
	body, _ := json.Marshal(payload)
	_, _ = f.Write(append(body, '\n'))
}

func appendRecoveryDeliveryFailureAlert(challenge *authRecoveryChallenge, user *sysUser, channel, target string, err error) {
	if challenge == nil {
		return
	}
	payload := map[string]any{
		"challengeId":     challenge.ID,
		"userId":          challenge.UserID,
		"username":        challenge.Username,
		"channel":         channel,
		"targetMasked":    maskRecoveryTarget(target, channel),
		"error":           errString(err),
		"challengeStatus": challenge.Status,
		"expiresAt":       challenge.ExpiresAt.Format(time.RFC3339),
	}
	if user != nil {
		payload["roleCode"] = user.RoleCode
		if user.OrganizationID != nil {
			payload["organizationId"] = *user.OrganizationID
		}
		if user.CampusID != nil {
			payload["campusId"] = *user.CampusID
		}
	}
	appendRecoveryAlertRecord(payload)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func encodeRecoveryAlertPreview(payload map[string]any) string {
	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(payload)
	return strings.TrimSpace(buf.String())
}
