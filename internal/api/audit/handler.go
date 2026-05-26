package audit

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"time"

	"edu-schedule-system/internal/code"
	pkgAudit "edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"

	"go.uber.org/zap"
)

type Handler struct {
	logger        *zap.Logger
	service       pkgAudit.QueryService
	targetService pkgAudit.TargetSearchService
}

func New(logger *zap.Logger, service pkgAudit.QueryService, targetService ...pkgAudit.TargetSearchService) *Handler {
	var targets pkgAudit.TargetSearchService
	if len(targetService) > 0 {
		targets = targetService[0]
	}
	return &Handler{logger: logger, service: service, targetService: targets}
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
		var query pkgAudit.Query
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

func (h *Handler) Meta() core.HandlerFunc {
	return func(ctx core.Context) {
		ctx.Payload(h.service.Meta())
	}
}

func (h *Handler) SearchTargets() core.HandlerFunc {
	return func(ctx core.Context) {
		var query pkgAudit.TargetSearchQuery
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if h.targetService == nil {
			ctx.Payload(&pkgAudit.TargetSearchResponse{Items: []pkgAudit.TargetOption{}})
			return
		}
		resp, err := h.targetService.Search(ctx.Request().Context(), ctx.SessionUserInfo(), query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

func (h *Handler) Export() core.HandlerFunc {
	return func(ctx core.Context) {
		var query pkgAudit.Query
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		rows, err := h.service.Export(query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		filename := fmt.Sprintf("audit-logs-%s.csv", time.Now().Format("20060102-150405"))
		ctx.SetHeader("Content-Type", "text/csv; charset=utf-8")
		ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		writer := csv.NewWriter(ctx.ResponseWriter())
		if err := pkgAudit.WriteCSV(writer, rows); err != nil {
			ctx.AbortWithError(core.Error(http.StatusInternalServerError, code.ServerError, code.Text(code.ServerError)))
			return
		}
		ctx.ResponseWriter().WriteHeader(http.StatusOK)
	}
}

func RegisterRoutes(h *Handler, listGroup core.RouterGroup, exportGroup core.RouterGroup) {
	listGroup.GET("/audit/logs", h.List())
	listGroup.GET("/audit/logs/meta", h.Meta())
	listGroup.GET("/audit/logs/targets", h.SearchTargets())
	exportGroup.GET("/audit/logs/export", h.Export())
}
