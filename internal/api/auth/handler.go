package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"
	authService "edu-schedule-system/internal/service/auth"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	logger  *zap.Logger
	service authService.Service
}

func New(logger *zap.Logger, service authService.Service) *Handler {
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
		case apperr.KindConflict:
			ctx.AbortWithError(core.Error(http.StatusConflict, code.Conflict, appErr.Message))
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

// Login godoc
// @Summary 正式账号密码登录
// @Description 使用 sys_user 用户表进行正式登录，签发 JWT。
// @Tags Auth
// @Accept json
// @Produce json
// @Param RequestBody body authService.LoginRequest true "请求参数"
// @Success 200 {object} authService.LoginResponse
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Router /api/v1/auth/login [post]
func (h *Handler) Login() core.HandlerFunc {
	return func(ctx core.Context) {
		var req authService.LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.Login(ctx.RequestContext(), &req, clientMetaFromRequest(ctx.Request()))
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.login", Module: "auth", TargetID: resp.UserInfo.ID, Detail: audit.Detail("username", resp.UserInfo.Username, "sessionId", sessionIDFromResponse(resp), "riskFlags", riskFlagsFromResponse(resp))})
		ctx.Payload(resp)
	}
}

// Refresh godoc
// @Summary 刷新访问令牌
// @Description 使用 refresh token 刷新 access token，并复用当前会话。
// @Tags Auth
// @Accept json
// @Produce json
// @Param RequestBody body authService.RefreshRequest true "请求参数"
// @Success 200 {object} authService.LoginResponse
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh() core.HandlerFunc {
	return func(ctx core.Context) {
		var req authService.RefreshRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.Refresh(ctx.RequestContext(), &req, clientMetaFromRequest(ctx.Request()))
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

// Me godoc
// @Summary 获取当前登录用户信息
// @Description 返回当前会话对应的用户信息。
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} authService.CurrentUser
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/auth/me [get]
func (h *Handler) Me() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.service.Me(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

// ChangePassword godoc
// @Summary 修改当前用户密码
// @Description 用于首次登录强制改密，也可用于当前登录用户主动修改密码。
// @Tags Auth
// @Accept json
// @Produce json
// @Param RequestBody body authService.ChangePasswordRequest true "请求参数"
// @Success 200 {object} authService.CurrentUser
// @Failure 400 {object} code.Failure
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/auth/change-password [post]
func (h *Handler) ChangePassword() core.HandlerFunc {
	return func(ctx core.Context) {
		var req authService.ChangePasswordRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		requestCtx := context.WithValue(ctx.RequestContext(), authService.RawAccessTokenContextKey{}, bearerToken(ctx.GetHeader("Authorization")))
		resp, err := h.service.ChangePassword(requestCtx, ctx.SessionUserInfo(), &req)
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.change_password", Module: "auth", TargetID: ctx.SessionUserInfo().Id})
		ctx.Payload(resp)
	}
}

// Logout godoc
// @Summary 当前用户登出
// @Description 撤销当前 access token，并在提供 refresh token 时一并使其失效。
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Param X-Refresh-Token header string false "可选 refresh token"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout() core.HandlerFunc {
	return func(ctx core.Context) {
		authorization := bearerToken(ctx.GetHeader("Authorization"))
		refreshToken := strings.TrimSpace(ctx.GetHeader("X-Refresh-Token"))
		if err := h.service.Logout(ctx.RequestContext(), ctx.SessionUserInfo(), authorization, refreshToken); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.logout", Module: "auth", TargetID: ctx.SessionUserInfo().Id})
		ctx.Payload(map[string]any{"success": true})
	}
}

// ListSessions godoc
// @Summary 列出当前用户会话
// @Description 返回当前用户的可见登录会话列表。
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} authService.SessionListResponse
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/auth/sessions [get]
func (h *Handler) ListSessions() core.HandlerFunc {
	return func(ctx core.Context) {
		resp, err := h.service.ListSessions(ctx.RequestContext(), ctx.SessionUserInfo())
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		ctx.Payload(resp)
	}
}

// RevokeSession godoc
// @Summary 撤销指定会话
// @Description 撤销当前用户名下的指定会话。
// @Tags Auth
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} code.Failure
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/auth/sessions/{id} [delete]
func (h *Handler) RevokeSession() core.HandlerFunc {
	return func(ctx core.Context) {
		sessionID := strings.TrimSpace(ctx.Param("id"))
		if sessionID == "" {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, "session id is required"))
			return
		}
		if err := h.service.RevokeSession(ctx.RequestContext(), ctx.SessionUserInfo(), sessionID, bearerToken(ctx.GetHeader("Authorization"))); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.session_revoke", Module: "auth", TargetID: ctx.SessionUserInfo().Id, Detail: audit.Detail("sessionId", sessionID)})
		ctx.Payload(map[string]any{"success": true})
	}
}

// LogoutAll godoc
// @Summary 注销全部其他会话
// @Description 撤销当前用户的全部会话，包含当前 access token 对应会话。
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Security BearerAuth
// @Router /api/v1/auth/logout-all [post]
func (h *Handler) LogoutAll() core.HandlerFunc {
	return func(ctx core.Context) {
		if err := h.service.LogoutAll(ctx.RequestContext(), ctx.SessionUserInfo(), bearerToken(ctx.GetHeader("Authorization"))); err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.logout_all", Module: "auth", TargetID: ctx.SessionUserInfo().Id})
		ctx.Payload(map[string]any{"success": true})
	}
}

// StartPasswordRecovery godoc
// @Summary 发起密码恢复
// @Description 通过预设恢复渠道发起找回密码挑战。
// @Tags Auth
// @Accept json
// @Produce json
// @Param RequestBody body authService.PasswordRecoveryStartRequest true "请求参数"
// @Success 200 {object} authService.PasswordRecoveryStartResponse
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Router /api/v1/auth/password-recovery/start [post]
func (h *Handler) StartPasswordRecovery() core.HandlerFunc {
	return func(ctx core.Context) {
		var req authService.PasswordRecoveryStartRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.StartPasswordRecovery(ctx.RequestContext(), &req, clientMetaFromRequest(ctx.Request()))
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.password_recovery_start", Module: "auth", Detail: audit.Detail("username", req.Username, "channel", req.Channel, "secondFactorChannel", req.SecondFactorChannel, "challengeId", resp.ChallengeID, "riskFlags", resp.RiskFlags)})
		ctx.Payload(resp)
	}
}

// ResetPasswordByRecovery godoc
// @Summary 完成密码恢复并重置密码
// @Description 校验恢复挑战后重置密码，并撤销历史会话。
// @Tags Auth
// @Accept json
// @Produce json
// @Param RequestBody body authService.PasswordRecoveryResetRequest true "请求参数"
// @Success 200 {object} authService.PasswordRecoveryResetResponse
// @Failure 400 {object} code.Failure
// @Failure 403 {object} code.Failure
// @Router /api/v1/auth/password-recovery/reset [post]
func (h *Handler) ResetPasswordByRecovery() core.HandlerFunc {
	return func(ctx core.Context) {
		var req authService.PasswordRecoveryResetRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.AbortWithError(core.Error(http.StatusBadRequest, code.ParamBindError, err.Error()))
			return
		}
		resp, err := h.service.ResetPasswordByRecovery(ctx.RequestContext(), &req, clientMetaFromRequest(ctx.Request()))
		if err != nil {
			h.abortServiceError(ctx, err)
			return
		}
		audit.Log(ctx, audit.Event{Action: "auth.password_recovery_complete", Module: "auth", Detail: audit.Detail("challengeId", req.ChallengeID, "revokedSessionCount", resp.RevokedSessionCount)})
		ctx.Payload(resp)
	}
}

func RegisterRoutes(h *Handler, publicGroup core.RouterGroup, securedGroup core.RouterGroup) {
	loginGroup := publicGroup.Group("", authRateLimitMiddleware(h.logger, "login", 10, time.Minute, "ip"))
	refreshGroup := publicGroup.Group("", authRateLimitMiddleware(h.logger, "refresh", 30, time.Minute, "ip"))
	recoveryStartGroup := publicGroup.Group("", authRateLimitMiddleware(h.logger, "recovery_start", 5, 10*time.Minute, "ip"))
	recoveryResetGroup := publicGroup.Group("", authRateLimitMiddleware(h.logger, "recovery_reset", 5, 10*time.Minute, "ip"))
	changePasswordGroup := securedGroup.Group("", authRateLimitMiddleware(h.logger, "change_password", 10, 10*time.Minute, "token"))

	loginGroup.POST("/auth/login", h.Login())
	refreshGroup.POST("/auth/refresh", h.Refresh())
	recoveryStartGroup.POST("/auth/password-recovery/start", h.StartPasswordRecovery())
	recoveryResetGroup.POST("/auth/password-recovery/reset", h.ResetPasswordByRecovery())
	securedGroup.GET("/auth/me", h.Me())
	changePasswordGroup.POST("/auth/change-password", h.ChangePassword())
	securedGroup.POST("/auth/logout", h.Logout())
	securedGroup.GET("/auth/sessions", h.ListSessions())
	securedGroup.DELETE("/auth/sessions/:id", h.RevokeSession())
	securedGroup.POST("/auth/logout-all", h.LogoutAll())
}

func clientIPFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	for _, key := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if value := strings.TrimSpace(req.Header.Get(key)); value != "" {
			parts := strings.Split(value, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
			return value
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

func clientMetaFromRequest(req *http.Request) authService.AuthClientMeta {
	if req == nil {
		return authService.AuthClientMeta{}
	}
	return authService.AuthClientMeta{
		ClientIP:   clientIPFromRequest(req),
		UserAgent:  strings.TrimSpace(req.UserAgent()),
		DeviceName: strings.TrimSpace(req.Header.Get("X-Device-Name")),
	}
}

func bearerToken(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		value = strings.TrimSpace(value[7:])
	}
	return value
}

func sessionIDFromResponse(resp *authService.LoginResponse) string {
	if resp == nil || resp.Session == nil {
		return ""
	}
	return resp.Session.ID
}

func riskFlagsFromResponse(resp *authService.LoginResponse) []string {
	if resp == nil || resp.Session == nil {
		return nil
	}
	return resp.Session.RiskFlags
}
