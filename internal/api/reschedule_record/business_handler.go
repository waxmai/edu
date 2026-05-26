package reschedule_record

import (
	"errors"
	"net/http"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gorm"
)

// ListBusiness godoc
// @Summary 调补课记录列表
// @Description 按学员、操作类型和日期范围查询调补课记录。
// @Tags Business.reschedule_record
// @Accept json
// @Produce json
// @Param studentId query int false "学员ID"
// @Param operationType query string false "操作类型"
// @Param startDate query string false "开始日期，格式 2006-01-02"
// @Param endDate query string false "结束日期，格式 2006-01-02"
// @Success 200 {object} dto.RescheduleRecordListResponse
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/reschedule-records [get]
func (h *Handler) ListBusiness() core.HandlerFunc {
	return func(ctx core.Context) {
		query := dto.RescheduleRecordListQuery{}
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		if ctx.SessionUserInfo().DataScope == proposal.DataScopeSelf {
			ctx.Payload(dto.RescheduleRecordListResponse{})
			return
		}

		list, err := h.rescheduleRecordService.List(ctx.RequestContext(), query)
		if err != nil {
			h.abortBusinessError(ctx, err)
			return
		}

		ctx.Payload(list)
	}
}

func (h *Handler) abortBusinessError(ctx core.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithError(core.Error(http.StatusNotFound, code.RecordNotFound, "record not found"))
		return
	}
	h.abortServiceError(ctx, err)
}
