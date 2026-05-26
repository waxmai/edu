package schedule

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"
	"gorm.io/gorm"
)

type scheduleIDURI struct {
	ID int `uri:"id" binding:"required"`
}

// Cancel godoc
// @Summary 取消排课
// @Description 将指定排课取消，并写入取消原因。
// @Tags Business.schedule
// @Accept json
// @Produce json
// @Param id path int true "排课ID"
// @Param RequestBody body dto.ScheduleCancelRequest true "请求参数"
// @Success 200 {object} genResultInfo
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/schedules/{id}/cancel [patch]
func (h *Handler) Cancel() core.HandlerFunc {
	return func(ctx core.Context) {
		id, ok := parsePositiveID(ctx)
		if !ok {
			return
		}

		var req dto.ScheduleCancelRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		rowsAffected, err := h.scheduleService.Cancel(ctx.RequestContext(), int32(id), &req)
		if err != nil {
			h.handleBusinessError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "schedule.cancel", Module: "schedule", TargetID: int32(id), Detail: audit.Detail("reason", req.Reason)})

		ctx.Payload(&genResultInfo{RowsAffected: rowsAffected})
	}
}

// Leave godoc
// @Summary 学员请假
// @Description 将排课标记为请假，并生成调补课记录。
// @Tags Business.schedule
// @Accept json
// @Produce json
// @Param id path int true "排课ID"
// @Param RequestBody body dto.ScheduleLeaveRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/schedules/{id}/leave [post]
func (h *Handler) Leave() core.HandlerFunc {
	return func(ctx core.Context) {
		id, ok := parsePositiveID(ctx)
		if !ok {
			return
		}

		var req dto.ScheduleLeaveRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		recordID, err := h.scheduleService.Leave(ctx.RequestContext(), int32(id), &req)
		if err != nil {
			h.handleBusinessError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "schedule.leave", Module: "schedule", TargetID: int32(id), Detail: audit.Detail("rescheduleRecordId", recordID, "reason", req.Reason)})

		ctx.Payload(map[string]interface{}{"id": recordID})
	}
}

// Reschedule godoc
// @Summary 调课
// @Description 基于原排课创建新的排课时间，并生成调课记录。
// @Tags Business.schedule
// @Accept json
// @Produce json
// @Param id path int true "原排课ID"
// @Param RequestBody body dto.ScheduleRescheduleRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} code.Failure
// @Failure 409 {object} dto.ScheduleConflictResponse
// @Security BearerAuth
// @Router /api/v1/schedules/{id}/reschedule [post]
func (h *Handler) Reschedule() core.HandlerFunc {
	return func(ctx core.Context) {
		id, ok := parsePositiveID(ctx)
		if !ok {
			return
		}

		var req dto.ScheduleRescheduleRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		recordID, err := h.scheduleService.Reschedule(ctx.RequestContext(), int32(id), &req)
		if err != nil {
			h.handleBusinessError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "schedule.reschedule", Module: "schedule", TargetID: int32(id), Detail: audit.Detail("rescheduleRecordId", recordID, "newClassDate", req.NewClassDate, "newStartTime", req.NewStartTime, "newEndTime", req.NewEndTime)})

		ctx.Payload(map[string]interface{}{"id": recordID})
	}
}

// CreateMakeup godoc
// @Summary 创建补课排课
// @Description 基于原排课创建补课，并生成对应调补课记录。
// @Tags Business.schedule
// @Accept json
// @Produce json
// @Param RequestBody body dto.MakeupScheduleCreateRequest true "请求参数"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} code.Failure
// @Failure 409 {object} dto.ScheduleConflictResponse
// @Security BearerAuth
// @Router /api/v1/makeup-schedules [post]
func (h *Handler) CreateMakeup() core.HandlerFunc {
	return func(ctx core.Context) {
		var req dto.MakeupScheduleCreateRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}

		recordID, err := h.scheduleService.CreateMakeup(ctx.RequestContext(), &req)
		if err != nil {
			h.handleBusinessError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "schedule.makeup_create", Module: "schedule", TargetID: req.OriginalScheduleID, Detail: audit.Detail("rescheduleRecordId", recordID, "classDate", req.ClassDate, "startTime", req.StartTime, "endTime", req.EndTime)})

		ctx.Payload(map[string]interface{}{"id": recordID})
	}
}

func parsePositiveID(ctx core.Context) (int, bool) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		abortBadID(ctx)
		return 0, false
	}
	return id, true
}

func (h *Handler) handleBusinessError(ctx core.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.AbortWithError(core.Error(http.StatusNotFound, code.RecordNotFound, "record not found"))
		return
	}
	var appErr *apperr.Error
	if errors.As(err, &appErr) && appErr.Kind == apperr.KindConflict {
		teacherConflict, studentConflict := splitConflictMessage(err)
		if teacherConflict || studentConflict {
			ctx.PayloadWithStatus(http.StatusConflict, dto.ScheduleConflictResponse{TeacherConflict: teacherConflict, StudentConflict: studentConflict})
			return
		}
	}
	h.abortServiceError(ctx, err)
}

func splitConflictMessage(err error) (bool, bool) {
	if err == nil {
		return false, false
	}
	msg := err.Error()
	return strings.Contains(msg, "teacher=true"), strings.Contains(msg, "student=true")
}

func reqReason(reason string) string {
	return strings.TrimSpace(reason)
}
