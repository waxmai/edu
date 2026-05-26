package apierr

import "strings"

func TranslateMySQLConstraintMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return trimmed
	}
	if strings.Contains(trimmed, "chk_lesson_package_paid_not_exceed_total") {
		return "已缴金额不能超过总金额"
	}
	return trimmed
}
