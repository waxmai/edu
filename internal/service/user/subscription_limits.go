package user

import (
	"context"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
	"gorm.io/gorm"
)

func (s *service) ensureUserQuota(ctx context.Context, actor proposal.SessionUserInfo, organizationID int32) error {
	if actor.IsPlatformAdmin() || organizationID <= 0 {
		return nil
	}
	subscription, err := s.tenantService.GetSubscription(ctx, organizationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if subscription.Status == proposal.SubscriptionStatusExpired || subscription.Status == proposal.SubscriptionStatusPastDue || subscription.Status == proposal.SubscriptionStatusSuspended {
		return apperr.Forbidden("当前机构订阅不可用，无法新增账号")
	}
	if subscription.MaxUsers <= 0 {
		return nil
	}
	var count int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&sysUser{}).Where("organization_id = ?", organizationID).Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(subscription.MaxUsers) {
		return apperr.Forbidden("当前套餐账号数已达上限，请升级套餐")
	}
	return nil
}
