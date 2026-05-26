package auth

import (
	"context"
	"strings"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"gorm.io/gorm"
)

type authStore struct {
	db mysql.Repo
}

func newAuthStore(db mysql.Repo) *authStore {
	return &authStore{db: db}
}

func (s *authStore) findByUsername(ctx context.Context, username string) (*sysUser, error) {
	var user sysUser
	err := s.db.GetDbR().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *authStore) findByID(ctx context.Context, id int32) (*sysUser, error) {
	var user sysUser
	err := s.db.GetDbR().WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *authStore) findSessionByID(ctx context.Context, id string) (*authSession, error) {
	var row authSession
	if err := s.db.GetDbR().WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *authStore) findRecoveryChallengeByID(ctx context.Context, id string) (*authRecoveryChallenge, error) {
	var row authRecoveryChallenge
	if err := s.db.GetDbR().WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *authStore) recordFailedLogin(ctx context.Context, user *sysUser) error {
	if user == nil {
		return nil
	}
	attempts := user.FailedLoginCount + 1
	updates := map[string]interface{}{"failed_login_count": attempts}
	if attempts >= maxFailedLoginAttempts() {
		lockedUntil := time.Now().Add(lockDuration())
		updates["status"] = proposal.UserStatusLocked
		updates["locked_until"] = &lockedUntil
		user.Status = proposal.UserStatusLocked
		user.LockedUntil = &lockedUntil
	}
	user.FailedLoginCount = attempts
	return s.db.GetDbW().WithContext(ctx).Model(&sysUser{}).Where("id = ?", user.ID).Updates(updates).Error
}

func (s *authStore) clearFailedLoginState(ctx context.Context, userID int32, currentStatus string) error {
	updates := map[string]interface{}{
		"failed_login_count": 0,
		"locked_until":       nil,
	}
	if strings.TrimSpace(currentStatus) == proposal.UserStatusLocked {
		updates["status"] = proposal.UserStatusEnabled
	}
	return s.db.GetDbW().WithContext(ctx).Model(&sysUser{}).Where("id = ?", userID).Updates(updates).Error
}

func (s *authStore) revokeSessionByID(ctx context.Context, userID int32, sessionID, reason string) error {
	now := time.Now()
	result := s.db.GetDbW().WithContext(ctx).Model(&authSession{}).Where("id = ? AND user_id = ? AND status = ?", sessionID, userID, authSessionStatusActive).Updates(map[string]interface{}{
		"status":         authSessionStatusRevoked,
		"revoked_at":     now,
		"revoked_reason": strings.TrimSpace(reason),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *authStore) listSessions(ctx context.Context, userID int32) ([]authSession, error) {
	var rows []authSession
	if err := s.db.GetDbR().WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *authStore) listActiveSessions(ctx context.Context, userID int32) ([]authSession, error) {
	var sessions []authSession
	if err := s.db.GetDbR().WithContext(ctx).Where("user_id = ? AND status = ?", userID, authSessionStatusActive).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *authStore) countActiveSessions(ctx context.Context, userID int32) int {
	var count int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&authSession{}).Where("user_id = ? AND status = ?", userID, authSessionStatusActive).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (s *authStore) countSuspiciousSessions(ctx context.Context, userID int32) int64 {
	var suspiciousCount int64
	_ = s.db.GetDbR().WithContext(ctx).Model(&authSession{}).Where("user_id = ? AND status = ? AND is_suspicious = ?", userID, authSessionStatusActive, true).Count(&suspiciousCount).Error
	return suspiciousCount
}

func (s *authStore) revokeAllSessions(ctx context.Context, userID int32, reason string) error {
	now := time.Now()
	return s.db.GetDbW().WithContext(ctx).Model(&authSession{}).Where("user_id = ? AND status = ?", userID, authSessionStatusActive).Updates(map[string]interface{}{
		"status":         authSessionStatusRevoked,
		"revoked_at":     now,
		"revoked_reason": strings.TrimSpace(reason),
	}).Error
}

func (s *authStore) saveSession(ctx context.Context, sessionRow *authSession) error {
	if sessionRow == nil {
		return nil
	}
	return s.db.GetDbW().WithContext(ctx).Save(sessionRow).Error
}

func (s *authStore) createRecoveryChallengeWithAudit(ctx context.Context, challenge *authRecoveryChallenge, action, status string, detail map[string]any, meta AuthClientMeta) error {
	return s.db.GetDbW().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(challenge).Error; err != nil {
			return err
		}
		return insertRecoveryAuditTx(tx, challenge.ID, challenge.UserID, action, status, detail, meta)
	})
}

func (s *authStore) expireRecoveryChallenge(ctx context.Context, challengeID string) error {
	return s.db.GetDbW().WithContext(ctx).Model(&authRecoveryChallenge{}).Where("id = ?", challengeID).Updates(map[string]interface{}{
		"status": recoveryStatusExpired,
	}).Error
}

func (s *authStore) appendRecoveryAudit(ctx context.Context, challengeID string, userID int32, action, status string, detail map[string]any, meta AuthClientMeta) error {
	return insertRecoveryAuditTx(s.db.GetDbW().WithContext(ctx), challengeID, userID, action, status, detail, meta)
}
