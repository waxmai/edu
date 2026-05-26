package dashboard

import (
	"errors"
	"net/http"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"
	dashboardservice "edu-schedule-system/internal/service/dashboard"
	"edu-schedule-system/internal/service/dto"
)

type Handler struct {
	service dashboardservice.Service
}

func New(service dashboardservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) abortServiceError(ctx core.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case apperr.KindInvalidArgument:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		case apperr.KindForbidden:
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, appErr.Message))
		case apperr.KindNotFound:
			ctx.AbortWithError(core.Error(http.StatusNotFound, code.RecordNotFound, appErr.Message))
		default:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		}
		return
	}
	ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
}

func (h *Handler) Overview() core.HandlerFunc {
	return func(ctx core.Context) {
		var query dto.DashboardQuery
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.Overview(ctx.RequestContext(), ctx.SessionUserInfo(), query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func RegisterRoutes(h *Handler, group core.RouterGroup) {
	group.GET("/statistics/dashboard", h.Overview())
}
