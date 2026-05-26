package auth

import (
	"context"
	"time"

	"edu-schedule-system/internal/proposal"
	"gorm.io/gorm"
)

func (s *authStore) completePasswordRecovery(ctx context.Context, user *sysUser, challenge *authRecoveryChallenge, passwordHash string, meta AuthClientMeta) (int64, time.Time, error) {
	now := time.Now()
	var revokedCount int64
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
		result := tx.Model(&authSession{}).Where("user_id = ? AND status = ?", user.ID, authSessionStatusActive).Updates(map[string]interface{}{
			"status":               authSessionStatusRevoked,
			"revoked_at":           now,
			"revoked_reason":       "password_recovery_reset",
			"recovery_verified_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		revokedCount = result.RowsAffected
		if err := tx.Model(&authRecoveryChallenge{}).Where("id = ?", challenge.ID).Updates(map[string]interface{}{
			"status":      recoveryStatusConsumed,
			"verified_at": now,
			"consumed_at": now,
		}).Error; err != nil {
			return err
		}
		return insertRecoveryAuditTx(tx, challenge.ID, user.ID, "recovery.completed", "success", map[string]any{
			"revokedSessionCount": revokedCount,
			"channels":            []string{challenge.Channel, challenge.SecondChannel},
		}, meta)
	})
	return revokedCount, now, err
}
