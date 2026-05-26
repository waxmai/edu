package notification

import (
	"context"
	"errors"
	"net/http"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"
	notificationservice "edu-schedule-system/internal/service/notification"

	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	service notificationservice.QueryService
}

func New(logger *zap.Logger, service notificationservice.QueryService) *Handler {
	if service == nil {
		service = notificationservice.NoopQueryService{}
	}
	return &Handler{logger: logger, service: service}
}

func (h *Handler) List() core.HandlerFunc {
	return func(ctx core.Context) {
		var query notificationservice.Query
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.List(ctx.Request().Context(), query)
		if err != nil {
			h.abortError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) abortError(ctx core.Context, err error) {
	if err == nil {
		return
	}
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
	if errors.Is(err, context.Canceled) {
		ctx.AbortWithError(core.Error(http.StatusRequestTimeout, code.ServerError, err.Error()))
		return
	}
	ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
}

func RegisterRoutes(h *Handler, group core.RouterGroup) {
	group.GET("/notification-records", h.List())
}
