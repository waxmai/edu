package lesson_record

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
// @Summary 上课记录列表
// @Description 按业务查询条件返回上课记录列表，支持按学员、教师、出勤状态和日期范围筛选。
// @Tags Business.lesson_record
// @Accept json
// @Produce json
// @Param studentId query int false "学员ID"
// @Param teacherId query int false "教师ID"
// @Param attendanceStatus query string false "出勤状态"
// @Param startDate query string false "开始日期，格式 2006-01-02"
// @Param endDate query string false "结束日期，格式 2006-01-02"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} dto.LessonRecordListResponse
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/lesson-records [get]
func (h *Handler) ListBusiness() core.HandlerFunc {
	return func(ctx core.Context) {
		query := dto.LessonRecordListQuery{}
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
			query.TeacherID = ctx.SessionUserInfo().Id
		}

		list, err := h.lesson_recordService.List(ctx.RequestContext(), query)
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
