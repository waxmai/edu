package recoveryalert

import (
	"errors"
	"net/http"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	pkgRecoveryAlert "edu-schedule-system/internal/pkg/recoveryalert"
	"edu-schedule-system/internal/service/apperr"

	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service pkgRecoveryAlert.QueryService
}

func New(logger *zap.Logger, service pkgRecoveryAlert.QueryService) *Handler {
	return &Handler{logger: logger, service: service}
}

func (h *Handler) abortServiceError(ctx core.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		switch appErr.Kind {
		case apperr.KindInvalidArgument:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		case apperr.KindForbidden:
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, appErr.Message))
		default:
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, appErr.Message))
		}
		return
	}
	ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
}

func (h *Handler) List() core.HandlerFunc {
	return func(ctx core.Context) {
		var query pkgRecoveryAlert.Query
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.List(query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) Summary() core.HandlerFunc {
	return func(ctx core.Context) {
		var query pkgRecoveryAlert.Query
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.Summary(query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func RegisterRoutes(h *Handler, group core.RouterGroup) {
	group.GET("/recovery-alerts", h.List())
	group.GET("/recovery-alerts/summary", h.Summary())
}
