package tenant

import (
	"context"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"gorm.io/gorm"
)

func EnsureTenantWriteAllowed(ctx context.Context, svc Service, actor proposal.SessionUserInfo) error {
	if actor.IsPlatformAdmin() || actor.OrganizationID <= 0 {
		return nil
	}
	return EnsureTenantWriteAllowedForOrganization(ctx, svc, actor.OrganizationID)
}

func EnsureTenantWriteAllowedForOrganization(ctx context.Context, svc Service, organizationID int32) error {
	if organizationID <= 0 {
		return nil
	}
	base, ok := svc.(*service)
	if !ok || base == nil {
		return nil
	}
	var subscription model.Subscription
	if err := base.db.GetDbR().WithContext(ctx).Where("organization_id = ?", organizationID).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if !proposal.IsActiveSubscriptionStatus(subscription.Status) {
		return apperr.Forbidden("当前机构订阅不可用，已禁止关键写操作")
	}
	return nil
}
