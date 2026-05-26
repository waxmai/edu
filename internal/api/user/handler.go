package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"
	userService "edu-schedule-system/internal/service/user"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	logger  *zap.Logger
	service userService.Service
}

func New(logger *zap.Logger, service userService.Service) *Handler {
	return &Handler{logger: logger, service: service}
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

// List godoc
// @Summary 用户列表
// @Description 查询系统用户列表，支持平台/机构/校区分级管理视角。
// @Tags Users
// @Accept json
// @Produce json
// @Param username query string false "用户名"
// @Param realName query string false "姓名"
// @Param roleCode query string false "角色编码"
// @Param organizationId query int false "机构ID"
// @Param campusId query int false "校区ID"
// @Param status query string false "状态"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} userService.ListResponse
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users [get]
// List 用户列表
func (h *Handler) List() core.HandlerFunc {
	return func(ctx core.Context) {
		var query userService.ListQuery
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.List(ctx.RequestContext(), ctx.SessionUserInfo(), query)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

// Create godoc
// @Summary 新增用户
// @Description 新增 sys_user，支持平台/机构/校区分级管理。
// @Tags Users
// @Accept json
// @Produce json
// @Param RequestBody body userService.CreateRequest true "请求参数"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users [post]
// Create 新增用户
func (h *Handler) Create() core.HandlerFunc {
	return func(ctx core.Context) {
		var req userService.CreateRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		id, err := h.service.Create(ctx.RequestContext(), ctx.SessionUserInfo(), &req)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "user.create", Module: "user", TargetID: id, Detail: audit.Detail("username", req.Username, "roleCode", req.RoleCode, "status", req.Status)})
		ctx.Payload(map[string]any{"id": id})
	}
}

// Update godoc
// @Summary 修改用户
// @Description 修改 sys_user，支持平台/机构/校区分级管理。
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param RequestBody body userService.UpdateRequest true "请求参数"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
// Update 修改用户
func (h *Handler) Update() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req userService.UpdateRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.service.Update(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "user.update", Module: "user", TargetID: int32(id), Detail: audit.Detail("username", req.Username, "roleCode", req.RoleCode, "status", req.Status, "passwordReset", req.Password != "")})
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

// UpdateStatus godoc
// @Summary 修改用户状态
// @Description 修改 sys_user 状态，支持平台/机构/校区分级管理。
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param RequestBody body userService.UpdateStatusRequest true "请求参数"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users/{id}/status [patch]
// UpdateStatus 修改用户状态
func (h *Handler) UpdateStatus() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req userService.UpdateStatusRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.service.UpdateStatus(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "user.update_status", Module: "user", TargetID: int32(id), Detail: audit.Detail("status", req.Status)})
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

// ResetPassword godoc
// @Summary 重置用户密码
// @Description 重置指定用户密码，并强制其下次登录改密。必须显式提供新密码，支持平台/机构/校区分级管理。
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param RequestBody body userService.ResetPasswordRequest false "请求参数"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users/{id}/reset-password [post]
func (h *Handler) ResetPassword() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		var req userService.ResetPasswordRequest
		if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		if err := h.service.ResetPassword(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id), &req); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "user.reset_password", Module: "user", TargetID: int32(id), Detail: audit.Detail("passwordProvided", strings.TrimSpace(req.NewPassword) != "")})
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

// Unlock godoc
// @Summary 解锁用户
// @Description 清空失败登录计数并解除锁定，支持平台/机构/校区分级管理。
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/users/{id}/unlock [post]
func (h *Handler) Unlock() core.HandlerFunc {
	return func(ctx core.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil || id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "id must be a positive integer"))
			return
		}
		if err := h.service.Unlock(ctx.RequestContext(), ctx.SessionUserInfo(), int32(id)); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "user.unlock", Module: "user", TargetID: int32(id)})
		ctx.Payload(map[string]any{"rows_affected": 1})
	}
}

func RegisterRoutes(h *Handler, group core.RouterGroup) {
	group.GET("/users", h.List())
	group.POST("/users", h.Create())
	group.PUT("/users/:id", h.Update())
	group.PATCH("/users/:id/status", h.UpdateStatus())
	group.POST("/users/:id/reset-password", h.ResetPassword())
	group.POST("/users/:id/unlock", h.Unlock())
}
