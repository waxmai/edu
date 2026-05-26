package auth

import (
	"context"
	"time"

	"edu-schedule-system/internal/proposal"
	"gorm.io/gorm"
)

func (s *authStore) changePasswordAndRevokeOtherSessions(ctx context.Context, user *sysUser, passwordHash string, currentSessionID string) (time.Time, error) {
	now := time.Now()
	err := s.db.GetDbW().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&sysUser{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"password_hash":            passwordHash,
			"must_change_password":     false,
			"last_password_changed_at": now,
			"failed_login_count":       0,
			"locked_until":             nil,
			"status":                   proposal.UserStatusEnabled,
		}).Error; err != nil {
			return err
		}
		return revokeUserSessionsTx(tx, user.ID, currentSessionID, "password_changed")
	})
	return now, err
}
