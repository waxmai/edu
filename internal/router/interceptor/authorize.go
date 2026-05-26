package interceptor

import (
	"net/http"
	"strings"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
)

func (i *interceptor) RequireRoles(roles ...string) core.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			allowed[role] = struct{}{}
		}
	}

	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if !user.IsEnabled() {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "账号不可用"))
			return
		}
		if len(allowed) == 0 {
			return
		}
		if _, ok := allowed[user.RoleCode]; !ok {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "无权限访问该接口"))
			return
		}
		if !user.HasFeature(proposal.FeatureAuth) {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "当前租户未开通认证功能"))
			return
		}
	}
}

func (i *interceptor) RequireAdmin() core.HandlerFunc {
	return i.RequireRoles(proposal.RolePlatformAdmin, proposal.RoleOrgAdmin, proposal.RoleCampusAdmin)
}

func (i *interceptor) RequirePlatformRoles() core.HandlerFunc {
	return i.RequireRoles(proposal.RolePlatformAdmin, proposal.RolePlatformOps, proposal.RolePlatformFinance, proposal.RolePlatformAuditor, proposal.RolePlatformSupport)
}

func (i *interceptor) RequirePermissions(permissions ...string) core.HandlerFunc {
	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if !user.IsEnabled() {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "账号不可用"))
			return
		}
		if len(permissions) == 0 {
			return
		}
		if !user.HasAnyPermission(permissions...) {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "无权限访问该接口"))
			return
		}
	}
}

func (i *interceptor) RequireFeature(feature string) core.HandlerFunc {
	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if feature == "" || user.IsPlatformAdmin() {
			return
		}
		if !user.HasFeature(feature) {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "当前租户未开通该功能"))
			return
		}
	}
}

func (i *interceptor) RequireTenantActive() core.HandlerFunc {
	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if user.IsPlatformAdmin() {
			return
		}
		if user.SubscriptionStatus == proposal.SubscriptionStatusSuspended {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "当前机构已停用，请联系平台管理员"))
			return
		}
		if user.SubscriptionStatus == proposal.SubscriptionStatusExpired || user.SubscriptionStatus == proposal.SubscriptionStatusPastDue {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "当前机构订阅已过期，请续费后继续操作"))
			return
		}
	}
}

func (i *interceptor) RequireDataScopeSelfOrAll() core.HandlerFunc {
	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if !user.IsEnabled() {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "账号不可用"))
			return
		}
		if user.DataScope == proposal.DataScopeAll || user.DataScope == proposal.DataScopeOrg || user.DataScope == proposal.DataScopeCampus || user.DataScope == proposal.DataScopeSelf {
			return
		}
		ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "当前账号未配置可用数据范围"))
	}
}
