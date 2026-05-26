package tenant

import (
	"errors"
	"net/http"
	"strconv"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/router/interceptor"
	"edu-schedule-system/internal/service/apperr"
	tenantservice "edu-schedule-system/internal/service/tenant"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	logger            *zap.Logger
	management        tenantservice.ManagementService
	subscriptionQuery tenantservice.SubscriptionQueryService
}

func New(logger *zap.Logger, management tenantservice.ManagementService, subscriptionQuery tenantservice.SubscriptionQueryService) *Handler {
	return &Handler{logger: logger, management: management, subscriptionQuery: subscriptionQuery}
}

func (h *Handler) abortServiceError(ctx core.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case apperr.KindInvalidArgument:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		case apperr.KindConflict:
			ctx.AbortWithError(core.Error(http.StatusConflict, code.Conflict, appErr.Message))
		case apperr.KindForbidden:
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, appErr.Message))
		default:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		}
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithError(core.Error(http.StatusNotFound, code.RecordNotFound, "record not found"))
		return
	}
	ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
}

func (h *Handler) ListOrganizations() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.management.ListOrganizations(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) CreateOrganization() core.HandlerFunc {
	return func(ctx core.Context) {
		var req tenantservice.CreateOrganizationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		id, err := h.management.CreateOrganization(ctx.RequestContext(), ctx.SessionUserInfo(), &req)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"id": id})
	}
}

func (h *Handler) UpdateOrganization() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req tenantservice.UpdateOrganizationRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.management.UpdateOrganization(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

func (h *Handler) ListCampuses() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.management.ListCampuses(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) CreateCampus() core.HandlerFunc {
	return func(ctx core.Context) {
		var req tenantservice.CreateCampusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		id, err := h.management.CreateCampus(ctx.RequestContext(), ctx.SessionUserInfo(), &req)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"id": id})
	}
}

func (h *Handler) UpdateCampus() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req tenantservice.UpdateCampusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.management.UpdateCampus(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

func (h *Handler) GetOrganizationSettings() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.management.GetOrganizationSettings(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) UpdateOrganizationSettings() core.HandlerFunc {
	return func(ctx core.Context) {
		var req tenantservice.UpdateOrganizationSettingsRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.management.UpdateOrganizationSettings(ctx.RequestContext(), ctx.SessionUserInfo(), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

func (h *Handler) GetCurrentSubscription() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.subscriptionQuery.GetCurrentSubscription(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) ListSubscriptionPipeline() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.subscriptionQuery.ListSubscriptionPipeline(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) UpdateSubscriptionFollowUp() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req tenantservice.UpdateSubscriptionFollowUpRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.subscriptionQuery.UpdateSubscriptionFollowUp(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

func RegisterRoutes(h *Handler, tenantGroup core.RouterGroup, platformGroup core.RouterGroup, campusBaseGroup core.RouterGroup, authInterceptor interceptor.Interceptor) {
	tenantReadGroup := platformGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformTenantList, proposal.PermPlatformTenantGet))
	tenantWriteGroup := platformGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformTenantCreate, proposal.PermPlatformTenantUpdate, proposal.PermPlatformTenantDisable, proposal.PermPlatformTenantRestore))
	campusReadGroup := campusBaseGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformCampusList))
	campusWriteGroup := campusBaseGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformCampusCreate, proposal.PermPlatformCampusUpdate, proposal.PermPlatformCampusDisable))
	subscriptionReadGroup := platformGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformSubscriptionList, proposal.PermPlatformSubscriptionGet, proposal.PermPlatformCSList))
	subscriptionWriteGroup := platformGroup.Group("", authInterceptor.RequirePermissions(proposal.PermPlatformSubscriptionUpdate, proposal.PermPlatformSubscriptionQuota, proposal.PermPlatformSubscriptionTrial, proposal.PermPlatformSubscriptionSuspend, proposal.PermPlatformSubscriptionResume, proposal.PermPlatformCSUpdateOwner, proposal.PermPlatformCSAddFollowup, proposal.PermPlatformCSUpdateStatus))

	tenantReadGroup.GET("/platform/organizations", h.ListOrganizations())
	tenantWriteGroup.POST("/platform/organizations", h.CreateOrganization())
	tenantWriteGroup.PUT("/platform/organizations/:id", h.UpdateOrganization())
	campusReadGroup.GET("/platform/campuses", h.ListCampuses())
	campusWriteGroup.POST("/platform/campuses", h.CreateCampus())
	campusWriteGroup.PUT("/platform/campuses/:id", h.UpdateCampus())
	subscriptionReadGroup.GET("/platform/subscriptions", h.ListSubscriptionPipeline())
	subscriptionWriteGroup.PUT("/platform/subscriptions/:id/follow-up", h.UpdateSubscriptionFollowUp())
	tenantGroup.GET("/organization/settings", h.GetOrganizationSettings())
	tenantGroup.PUT("/organization/settings", h.UpdateOrganizationSettings())
	tenantGroup.GET("/subscriptions/current", h.GetCurrentSubscription())
}
