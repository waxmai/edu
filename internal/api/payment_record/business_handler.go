package payment_record

import (
	"errors"
	"net/http"
	"strconv"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gorm"
)

// ListBusiness godoc
// @Summary 缴费记录列表
// @Description 按业务查询条件返回缴费记录列表，支持按学员、课时包、缴费类型、状态、方式和日期范围筛选。
// @Tags Business.payment_record
// @Accept json
// @Produce json
// @Param studentId query int false "学员ID"
// @Param lessonPackageId query int false "课时包ID"
// @Param paymentType query string false "缴费类型"
// @Param paymentStatus query string false "缴费状态"
// @Param paymentMethod query string false "支付方式"
// @Param startDate query string false "开始日期，格式 2006-01-02"
// @Param endDate query string false "结束日期，格式 2006-01-02"
// @Success 200 {object} dto.PaymentRecordListResponse
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/payment-records [get]
func (h *Handler) ListBusiness() core.HandlerFunc {
	return func(ctx core.Context) {
		query := dto.PaymentRecordListQuery{}
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		if rawID := ctx.Param("id"); rawID != "" {
			id, err := strconv.Atoi(rawID)
			if err != nil || id <= 0 {
				abortBadID(ctx)
				return
			}
			query.StudentID = int32(id)
		}

		if ctx.SessionUserInfo().DataScope == proposal.DataScopeSelf {
			ctx.Payload(dto.PaymentRecordListResponse{})
			return
		}

		list, err := h.paymentRecordService.List(ctx.RequestContext(), query)
		if err != nil {
			h.abortBusinessError(ctx, err)
			return
		}

		ctx.Payload(list)
	}
}

// IncomeStatistics godoc
// @Summary 收入统计
// @Description 按日期范围聚合缴费记录收入统计。
// @Tags Business.payment_record
// @Accept json
// @Produce json
// @Param startDate query string false "开始日期，格式 2006-01-02"
// @Param endDate query string false "结束日期，格式 2006-01-02"
// @Param groupBy query string false "分组粒度"
// @Success 200 {object} dto.PaymentIncomeStatsResponse
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/statistics/income [get]
func (h *Handler) IncomeStatistics() core.HandlerFunc {
	return func(ctx core.Context) {
		query := dto.PaymentIncomeStatsQuery{}
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		stats, err := h.paymentRecordService.IncomeStatistics(ctx.RequestContext(), query)
		if err != nil {
			h.abortBusinessError(ctx, err)
			return
		}

		ctx.Payload(stats)
	}
}

func (h *Handler) abortBusinessError(ctx core.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithError(core.Error(http.StatusNotFound, code.RecordNotFound, "record not found"))
		return
	}
	h.abortServiceError(ctx, err)
}
