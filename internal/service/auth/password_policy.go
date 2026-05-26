package auth

import (
	"fmt"
	"unicode"

	"edu-schedule-system/internal/service/apperr"
)

const minPasswordLength = 12

func validatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return apperr.InvalidArgument(fmt.Sprintf("newPassword length must be at least %d characters", minPasswordLength))
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return apperr.InvalidArgument("password must include upper-case, lower-case, digit, and special character")
	}
	return nil
}
