package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/alert"
	"edu-schedule-system/internal/pkg/idgen"
	"edu-schedule-system/internal/pkg/jwtoken"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/notification"
	tenantservice "edu-schedule-system/internal/service/tenant"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	accessTokenTTL           = 2 * time.Hour
	refreshTokenTTL          = 7 * 24 * time.Hour
	defaultMaxFailedAttempts = int32(5)
	defaultTemporaryLockTTL  = 15 * time.Minute

	authSessionStatusActive  = "active"
	authSessionStatusRevoked = "revoked"

	recoveryStatusPending  = "pending"
	recoveryStatusVerified = "verified"
	recoveryStatusConsumed = "consumed"
	recoveryStatusExpired  = "expired"

	defaultRecoveryCodeTTL      = 10 * time.Minute
	recoveryAuditDetailMaxBytes = 2048
)

type Service interface {
	Login(ctx context.Context, req *LoginRequest, meta AuthClientMeta) (*LoginResponse, error)
	Refresh(ctx context.Context, req *RefreshRequest, meta AuthClientMeta) (*LoginResponse, error)
	Me(ctx context.Context, session proposal.SessionUserInfo) (*CurrentUser, error)
	ChangePassword(ctx context.Context, session proposal.SessionUserInfo, req *ChangePasswordRequest) (*LoginResponse, error)
	Logout(ctx context.Context, session proposal.SessionUserInfo, rawAccessToken, rawRefreshToken string) error
	ListSessions(ctx context.Context, session proposal.SessionUserInfo) (*SessionListResponse, error)
	RevokeSession(ctx context.Context, session proposal.SessionUserInfo, targetSessionID, rawCurrentAccessToken string) error
	LogoutAll(ctx context.Context, session proposal.SessionUserInfo, rawCurrentAccessToken string) error
	StartPasswordRecovery(ctx context.Context, req *PasswordRecoveryStartRequest, meta AuthClientMeta) (*PasswordRecoveryStartResponse, error)
	ResetPasswordByRecovery(ctx context.Context, req *PasswordRecoveryResetRequest, meta AuthClientMeta) (*PasswordRecoveryResetResponse, error)
}

type service struct {
	db                mysql.Repo
	store             *authStore
	tokenRevoker      TokenRevoker
	tenantService     tenantservice.Service
	recoveryDelivery  RecoveryDeliverySender
	loginAlertMonitor *alert.LoginFailureMonitor
}

type AuthClientMeta struct {
	ClientIP   string
	UserAgent  string
	DeviceName string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type CurrentUser struct {
	ID                 int32    `json:"id"`
	Username           string   `json:"username"`
	RealName           string   `json:"realName"`
	RoleCode           string   `json:"roleCode"`
	OrganizationID     int32    `json:"organizationId"`
	CampusID           int32    `json:"campusId"`
	OrganizationName   string   `json:"organizationName"`
	CampusName         string   `json:"campusName"`
	Phone              string   `json:"phone"`
	Email              string   `json:"email"`
	Status             string   `json:"status"`
	SubscriptionStatus string   `json:"subscriptionStatus"`
	FeatureFlags       []string `json:"featureFlags"`
	MustChangePassword bool     `json:"mustChangePassword"`
	Remark             string   `json:"remark"`
	Permissions        []string `json:"permissions"`
	MenuPermissions    []string `json:"menuPermissions"`
	DataScope          string   `json:"dataScope"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type SessionInfo struct {
	ID              string     `json:"id"`
	DeviceID        string     `json:"deviceId"`
	DeviceName      string     `json:"deviceName"`
	ClientIP        string     `json:"clientIP"`
	UserAgent       string     `json:"userAgent"`
	Status          string     `json:"status"`
	Current         bool       `json:"current"`
	Suspicious      bool       `json:"suspicious"`
	OrganizationID  int32      `json:"organizationId"`
	CampusID        int32      `json:"campusId"`
	RoleCode        string     `json:"roleCode"`
	RiskFlags       []string   `json:"riskFlags"`
	LastSeenAt      *time.Time `json:"lastSeenAt"`
	LastRefreshedAt *time.Time `json:"lastRefreshedAt"`
	RevokedAt       *time.Time `json:"revokedAt,omitempty"`
	RevokedReason   string     `json:"revokedReason,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type SessionListResponse struct {
	Items []SessionInfo `json:"items"`
	Total int           `json:"total"`
}

type PasswordRecoveryStartRequest struct {
	Username            string `json:"username" binding:"required"`
	Channel             string `json:"channel" binding:"required"`
	SecondFactorChannel string `json:"secondFactorChannel"`
}

type PasswordRecoveryStartResponse struct {
	Accepted             bool     `json:"accepted"`
	ChallengeID          string   `json:"challengeId,omitempty"`
	ExpiresIn            int64    `json:"expiresIn,omitempty"`
	PrimaryChannel       string   `json:"primaryChannel,omitempty"`
	PrimaryTargetMasked  string   `json:"primaryTargetMasked,omitempty"`
	SecondFactorRequired bool     `json:"secondFactorRequired"`
	SecondFactorChannel  string   `json:"secondFactorChannel,omitempty"`
	SecondTargetMasked   string   `json:"secondTargetMasked,omitempty"`
	RiskFlags            []string `json:"riskFlags,omitempty"`
	Message              string   `json:"message,omitempty"`
}

type PasswordRecoveryResetRequest struct {
	ChallengeID      string `json:"challengeId" binding:"required"`
	VerificationCode string `json:"verificationCode" binding:"required"`
	SecondFactorCode string `json:"secondFactorCode"`
	NewPassword      string `json:"newPassword" binding:"required"`
}

type PasswordRecoveryResetResponse struct {
	Success             bool      `json:"success"`
	RevokedSessionCount int64     `json:"revokedSessionCount"`
	RequiresFreshLogin  bool      `json:"requiresFreshLogin"`
	RecoveredAt         time.Time `json:"recoveredAt"`
	RecoveryChallengeID string    `json:"recoveryChallengeId"`
}

type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refreshToken"`
	TokenType    string       `json:"tokenType"`
	ExpiresIn    int64        `json:"expiresIn"`
	UserInfo     *CurrentUser `json:"userInfo"`
	Session      *SessionInfo `json:"session,omitempty"`
}

func New(db mysql.Repo, opts ...Option) Service {
	deliverySender := buildRecoveryDeliverySender(configs.Get().Auth.RecoveryDelivery.Mode)
	s := &service{db: db, store: newAuthStore(db), tokenRevoker: noopTokenRevoker{}, tenantService: tenantservice.New(db), recoveryDelivery: deliverySender}
	if adapter, ok := deliverySender.(*notificationRecoveryDeliverySender); ok {
		adapter.recorder = notification.NewDBRecorder(db)
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

type Option func(*service)

func WithTokenRevoker(revoker TokenRevoker) Option {
	return func(s *service) {
		if s == nil || revoker == nil {
			return
		}
		s.tokenRevoker = revoker
	}
}

func WithLoginFailureMonitor(monitor *alert.LoginFailureMonitor) Option {
	return func(s *service) {
		s.loginAlertMonitor = monitor
	}
}

func WithRecoveryDelivery(sender RecoveryDeliverySender) Option {
	return func(s *service) {
		if s == nil || sender == nil {
			return
		}
		s.recoveryDelivery = sender
	}
}

func (s *service) Login(ctx context.Context, req *LoginRequest, meta AuthClientMeta) (*LoginResponse, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("login request is required")
	}
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		return nil, apperr.InvalidArgument("username and password are required")
	}

	user, err := s.store.findByUsername(ctx, username)
	if err != nil {
		s.recordLoginFailureAlert(ctx, username, "user_not_found", meta)
		return nil, apperr.Forbidden("用户名或密码错误")
	}
	if err := ensureUserAvailable(user); err != nil {
		return nil, err
	}
	if err := ensureUserUnlocked(user); err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if recordErr := s.store.recordFailedLogin(ctx, user); recordErr != nil {
			return nil, recordErr
		}
		s.recordLoginFailureAlert(ctx, username, "bad_password", meta)
		return nil, apperr.Forbidden("用户名或密码错误")
	}
	if err := s.store.clearFailedLoginState(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	if user.Status == proposal.UserStatusLocked {
		user.Status = proposal.UserStatusEnabled
	}
	return s.issueTokenPair(ctx, user, meta, nil)
}

func (s *service) recordLoginFailureAlert(ctx context.Context, username, reason string, meta AuthClientMeta) {
	if s == nil || s.loginAlertMonitor == nil {
		return
	}
	s.loginAlertMonitor.Record(ctx, alert.LoginFailureEvent{Username: username, Reason: reason, ClientIP: meta.ClientIP, UserAgent: meta.UserAgent})
}

func (s *service) Refresh(ctx context.Context, req *RefreshRequest, meta AuthClientMeta) (*LoginResponse, error) {
	if req == nil || strings.TrimSpace(req.RefreshToken) == "" {
		return nil, apperr.InvalidArgument("refreshToken is required")
	}
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if s.tokenRevoker != nil {
		revoked, err := s.tokenRevoker.IsRevoked(ctx, refreshToken)
		if err != nil {
			return nil, err
		}
		if revoked {
			return nil, apperr.Forbidden("refresh token 已失效，请重新登录")
		}
	}
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, apperr.Forbidden("refresh token 无效，请重新登录")
	}
	kind, sessionID := ParseTokenKindAndSessionID(claims.TokenID)
	if kind != "refresh" || sessionID == "" {
		return nil, apperr.Forbidden("refresh token 无效，请重新登录")
	}
	storedSession, err := s.store.findSessionByID(ctx, sessionID)
	if err != nil {
		return nil, apperr.Forbidden("refresh token 无效，请重新登录")
	}
	if storedSession.UserID != claims.Id || storedSession.Status != authSessionStatusActive || storedSession.RevokedAt != nil {
		return nil, apperr.Forbidden("refresh token 已失效，请重新登录")
	}
	if storedSession.RefreshTokenID != claims.TokenID || storedSession.RefreshTokenHash != tokenDigest(refreshToken) {
		return nil, apperr.Forbidden("refresh token 已失效，请重新登录")
	}
	user, err := s.store.findByID(ctx, claims.Id)
	if err != nil {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	if err := ensureUserAvailable(user); err != nil {
		return nil, err
	}
	if err := ensureUserUnlocked(user); err != nil {
		return nil, err
	}
	effective, err := s.resolveEffectiveProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	if err := validateEffectiveProfile(user, effective); err != nil {
		return nil, err
	}
	if err := s.tokenRevoker.Revoke(ctx, refreshToken, refreshTokenTTL); err != nil {
		return nil, err
	}
	if meta.ClientIP == "" {
		meta.ClientIP = storedSession.ClientIP
	}
	if meta.UserAgent == "" {
		meta.UserAgent = storedSession.UserAgent
	}
	if meta.DeviceName == "" {
		meta.DeviceName = storedSession.DeviceName
	}
	return s.issueTokenPair(ctx, user, meta, storedSession)
}

func (s *service) Me(ctx context.Context, session proposal.SessionUserInfo) (*CurrentUser, error) {
	if session.Id <= 0 {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	user, err := s.store.findByID(ctx, session.Id)
	if err != nil {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	if err := ensureUserAvailable(user); err != nil {
		return nil, err
	}
	if err := ensureUserUnlocked(user); err != nil {
		return nil, err
	}
	effective, err := s.resolveEffectiveProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	if err := validateEffectiveProfile(user, effective); err != nil {
		return nil, err
	}
	if sessionID := ExtractSessionID(session.TokenID); sessionID != "" {
		now := time.Now()
		_ = s.db.GetDbW().WithContext(ctx).Model(&authSession{}).Where("id = ? AND user_id = ? AND status = ?", sessionID, session.Id, authSessionStatusActive).Updates(map[string]interface{}{
			"last_seen_at": now,
		}).Error
	}
	return toCurrentUser(user, effective), nil
}

func (s *service) ChangePassword(ctx context.Context, session proposal.SessionUserInfo, req *ChangePasswordRequest) (*LoginResponse, error) {
	if session.Id <= 0 {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	if req == nil {
		return nil, apperr.InvalidArgument("change password request is required")
	}
	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPassword == "" || newPassword == "" {
		return nil, apperr.InvalidArgument("oldPassword and newPassword are required")
	}
	if len(newPassword) < minPasswordLength {
		return nil, apperr.InvalidArgument(fmt.Sprintf("newPassword length must be at least %d characters", minPasswordLength))
	}
	if err := validatePasswordStrength(newPassword); err != nil {
		return nil, err
	}
	if oldPassword == newPassword {
		return nil, apperr.InvalidArgument("new password must be different from old password")
	}

	var user sysUser
	if err := s.db.GetDbW().WithContext(ctx).First(&user, session.Id).Error; err != nil {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	if err := ensureUserAvailable(&user); err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return nil, apperr.Forbidden("原密码错误")
	}
	effective, err := s.resolveEffectiveProfile(ctx, &user)
	if err != nil {
		return nil, err
	}
	if err := validateEffectiveProfile(&user, effective); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	currentSessionID := ExtractSessionID(session.TokenID)
	now, err := s.store.changePasswordAndRevokeOtherSessions(ctx, &user, string(hash), currentSessionID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hash)
	user.MustChangePassword = false
	user.LastPasswordChangedAt = &now
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	user.Status = proposal.UserStatusEnabled

	var currentSession *authSession
	if currentSessionID != "" {
		if stored, findErr := s.store.findSessionByID(ctx, currentSessionID); findErr == nil {
			currentSession = stored
		}
	}
	resp, err := s.issueTokenPair(ctx, &user, AuthClientMeta{}, currentSession)
	if err != nil {
		return nil, err
	}
	if rawCurrentAccessToken := extractRawAccessToken(ctx); rawCurrentAccessToken != "" && s.tokenRevoker != nil {
		if revokeErr := s.tokenRevoker.Revoke(ctx, rawCurrentAccessToken, logoutTokenTTL); revokeErr != nil {
			return nil, revokeErr
		}
	}
	return resp, nil
}

func (s *service) Logout(ctx context.Context, session proposal.SessionUserInfo, rawAccessToken, rawRefreshToken string) error {
	if session.Id <= 0 {
		return apperr.Forbidden("未登录或登录失效")
	}
	rawAccessToken = strings.TrimSpace(rawAccessToken)
	if rawAccessToken == "" {
		return apperr.InvalidArgument("token is required")
	}
	if err := s.tokenRevoker.Revoke(ctx, rawAccessToken, logoutTokenTTL); err != nil {
		return err
	}
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken != "" {
		if err := s.tokenRevoker.Revoke(ctx, rawRefreshToken, refreshTokenTTL); err != nil {
			return err
		}
	}
	if sessionID := ExtractSessionID(session.TokenID); sessionID != "" {
		return s.store.revokeSessionByID(ctx, session.Id, sessionID, "user_logout")
	}
	return nil
}

func (s *service) ListSessions(ctx context.Context, session proposal.SessionUserInfo) (*SessionListResponse, error) {
	if session.Id <= 0 {
		return nil, apperr.Forbidden("未登录或登录失效")
	}
	rows, err := s.store.listSessions(ctx, session.Id)
	if err != nil {
		return nil, err
	}
	currentSessionID := ExtractSessionID(session.TokenID)
	items := make([]SessionInfo, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSessionInfo(&row, row.ID == currentSessionID))
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Current != items[j].Current {
			return items[i].Current
		}
		if items[i].Status != items[j].Status {
			return items[i].Status == authSessionStatusActive
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return &SessionListResponse{Items: items, Total: len(items)}, nil
}

func (s *service) RevokeSession(ctx context.Context, session proposal.SessionUserInfo, targetSessionID, rawCurrentAccessToken string) error {
	if session.Id <= 0 {
		return apperr.Forbidden("未登录或登录失效")
	}
	targetSessionID = strings.TrimSpace(targetSessionID)
	if targetSessionID == "" {
		return apperr.InvalidArgument("session id is required")
	}
	if err := s.store.revokeSessionByID(ctx, session.Id, targetSessionID, "user_revoke_session"); err != nil {
		return err
	}
	if targetSessionID == ExtractSessionID(session.TokenID) {
		rawCurrentAccessToken = strings.TrimSpace(rawCurrentAccessToken)
		if rawCurrentAccessToken != "" {
			_ = s.tokenRevoker.Revoke(ctx, rawCurrentAccessToken, logoutTokenTTL)
		}
	}
	return nil
}

func (s *service) LogoutAll(ctx context.Context, session proposal.SessionUserInfo, rawCurrentAccessToken string) error {
	if session.Id <= 0 {
		return apperr.Forbidden("未登录或登录失效")
	}
	if err := s.store.revokeAllSessions(ctx, session.Id, "user_logout_all"); err != nil {
		return err
	}
	rawCurrentAccessToken = strings.TrimSpace(rawCurrentAccessToken)
	if rawCurrentAccessToken != "" {
		_ = s.tokenRevoker.Revoke(ctx, rawCurrentAccessToken, logoutTokenTTL)
	}
	return nil
}

func (s *service) StartPasswordRecovery(ctx context.Context, req *PasswordRecoveryStartRequest, meta AuthClientMeta) (*PasswordRecoveryStartResponse, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("password recovery request is required")
	}
	username := strings.TrimSpace(req.Username)
	channel := normalizeRecoveryChannel(req.Channel)
	secondChannel := normalizeRecoveryChannel(req.SecondFactorChannel)
	if username == "" || channel == "" {
		return nil, apperr.InvalidArgument("username and channel are required")
	}
	if secondChannel != "" && secondChannel == channel {
		return nil, apperr.InvalidArgument("second factor channel must be different from primary channel")
	}

	user, err := s.store.findByUsername(ctx, username)
	if err != nil {
		return &PasswordRecoveryStartResponse{Accepted: true, Message: "如果账号存在，验证码已发送"}, nil
	}
	if user.Status == proposal.UserStatusDisabled {
		return nil, apperr.Forbidden("账号已禁用，无法找回密码")
	}

	primaryTarget := recoveryTarget(user, channel)
	if primaryTarget == "" {
		return nil, apperr.InvalidArgument("所选验证方式未配置")
	}
	if secondChannel == "" && shouldRequireSecondFactor(user, meta, s.activeSessionCount(ctx, user.ID)) {
		secondChannel = defaultSecondFactorChannel(user, channel)
	}
	secondaryTarget := ""
	if secondChannel != "" {
		secondaryTarget = recoveryTarget(user, secondChannel)
		if secondaryTarget == "" {
			return nil, apperr.InvalidArgument("二次验证方式未配置")
		}
	}
	primaryCode, err := generateNumericCode(6)
	if err != nil {
		return nil, err
	}
	secondaryCode := ""
	if secondChannel != "" {
		secondaryCode, err = generateNumericCode(6)
		if err != nil {
			return nil, err
		}
	}
	riskFlags := s.recoveryRiskFlags(ctx, user.ID, meta)
	now := time.Now()
	challenge := &authRecoveryChallenge{
		ID:                  idgen.GenerateUniqueID(),
		UserID:              user.ID,
		Username:            user.Username,
		Channel:             channel,
		ChannelTarget:       primaryTarget,
		CodeHash:            tokenDigest(primaryCode),
		SecondChannel:       secondChannel,
		SecondChannelTarget: secondaryTarget,
		SecondCodeHash:      tokenDigest(secondaryCode),
		Status:              recoveryStatusPending,
		RiskFlags:           marshalStringSlice(riskFlags),
		RequestedIP:         sanitizeClientIP(meta.ClientIP),
		RequestedUserAgent:  sanitizeUserAgent(meta.UserAgent),
		ExpiresAt:           now.Add(defaultRecoveryCodeTTL),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.store.createRecoveryChallengeWithAudit(ctx, challenge, "recovery.requested", "accepted", map[string]any{
		"channel":             channel,
		"secondFactorChannel": secondChannel,
		"riskFlags":           riskFlags,
		"deviceName":          normalizeDeviceName(meta),
	}, meta); err != nil {
		return nil, err
	}
	if err := s.sendRecoveryCode(ctx, challenge, user, channel, primaryTarget, primaryCode); err != nil {
		masked := maskRecoveryTarget(primaryTarget, channel)
		_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.delivery_failed", "error", map[string]any{"channel": channel, "targetMasked": masked, "reason": err.Error(), "challengeExpired": true}, meta)
		_ = s.store.expireRecoveryChallenge(ctx, challenge.ID)
		challenge.Status = recoveryStatusExpired
		appendRecoveryDeliveryFailureAlert(challenge, user, channel, primaryTarget, err)
		return nil, err
	}
	if secondChannel != "" {
		if err := s.sendRecoveryCode(ctx, challenge, user, secondChannel, secondaryTarget, secondaryCode); err != nil {
			masked := maskRecoveryTarget(secondaryTarget, secondChannel)
			_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.delivery_failed", "error", map[string]any{"channel": secondChannel, "targetMasked": masked, "reason": err.Error(), "challengeExpired": true}, meta)
			_ = s.store.expireRecoveryChallenge(ctx, challenge.ID)
			challenge.Status = recoveryStatusExpired
			appendRecoveryDeliveryFailureAlert(challenge, user, secondChannel, secondaryTarget, err)
			return nil, err
		}
	}
	resp := &PasswordRecoveryStartResponse{
		Accepted:             true,
		ChallengeID:          challenge.ID,
		ExpiresIn:            int64(defaultRecoveryCodeTTL.Seconds()),
		PrimaryChannel:       channel,
		PrimaryTargetMasked:  maskRecoveryTarget(primaryTarget, channel),
		SecondFactorRequired: secondChannel != "",
		SecondFactorChannel:  secondChannel,
		SecondTargetMasked:   maskRecoveryTarget(secondaryTarget, secondChannel),
		RiskFlags:            riskFlags,
		Message:              "验证码已创建，请通过正式恢复通道完成验证后重置密码",
	}
	if recoveryPreviewEnabled() {
		resp.Message = "验证码已创建，当前环境允许恢复预览，请仅在本地调试中使用。当前版本不再通过接口直接返回验证码。"
		_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.preview_enabled", "dev_only", map[string]any{
			"channel":             channel,
			"secondFactorChannel": secondChannel,
		}, meta)
	}
	return resp, nil
}

func (s *service) ResetPasswordByRecovery(ctx context.Context, req *PasswordRecoveryResetRequest, meta AuthClientMeta) (*PasswordRecoveryResetResponse, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("password recovery reset request is required")
	}
	challengeID := strings.TrimSpace(req.ChallengeID)
	primaryCode := strings.TrimSpace(req.VerificationCode)
	secondaryCode := strings.TrimSpace(req.SecondFactorCode)
	newPassword := strings.TrimSpace(req.NewPassword)
	if challengeID == "" || primaryCode == "" || newPassword == "" {
		return nil, apperr.InvalidArgument("challengeId, verificationCode, and newPassword are required")
	}
	if len(newPassword) < minPasswordLength {
		return nil, apperr.InvalidArgument(fmt.Sprintf("newPassword length must be at least %d characters", minPasswordLength))
	}
	if err := validatePasswordStrength(newPassword); err != nil {
		return nil, err
	}

	challenge, err := s.store.findRecoveryChallengeByID(ctx, challengeID)
	if err != nil {
		return nil, apperr.Forbidden("恢复挑战不存在或已失效")
	}
	if challenge.Status != recoveryStatusPending {
		return nil, apperr.Forbidden("恢复挑战已失效")
	}
	if time.Now().After(challenge.ExpiresAt) {
		_ = s.store.expireRecoveryChallenge(ctx, challenge.ID)
		return nil, apperr.Forbidden("验证码已过期，请重新发起找回密码")
	}
	if challenge.CodeHash != tokenDigest(primaryCode) {
		_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.verify_failed", "rejected", map[string]any{"reason": "primary_code_mismatch"}, meta)
		return nil, apperr.Forbidden("验证码错误")
	}
	if strings.TrimSpace(challenge.SecondChannel) != "" {
		if secondaryCode == "" || challenge.SecondCodeHash != tokenDigest(secondaryCode) {
			_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.verify_failed", "rejected", map[string]any{"reason": "second_factor_mismatch"}, meta)
			return nil, apperr.Forbidden("二次验证码错误")
		}
	}

	user, err := s.store.findByID(ctx, challenge.UserID)
	if err != nil {
		return nil, apperr.Forbidden("账号不存在或已失效")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	revokedCount, recoveredAt, err := s.store.completePasswordRecovery(ctx, user, challenge, string(hash), meta)
	if err != nil {
		return nil, err
	}
	return &PasswordRecoveryResetResponse{
		Success:             true,
		RevokedSessionCount: revokedCount,
		RequiresFreshLogin:  true,
		RecoveredAt:         recoveredAt,
		RecoveryChallengeID: challenge.ID,
	}, nil
}

func (s *service) issueTokenPair(ctx context.Context, user *sysUser, meta AuthClientMeta, existingSession *authSession) (*LoginResponse, error) {
	jwtCfg := configs.Get().JWT
	now := time.Now()
	effective, err := s.resolveEffectiveProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	if err := validateEffectiveProfile(user, effective); err != nil {
		return nil, err
	}
	sessionRow := existingSession
	if sessionRow == nil {
		sessionRow = &authSession{ID: idgen.GenerateUniqueID(), UserID: user.ID, CreatedAt: now}
	}
	sessionRow.OrganizationID = nullableInt32(effective.OrganizationID)
	sessionRow.CampusID = nullableInt32(effective.CampusID)
	sessionRow.RoleCode = user.RoleCode
	applyClientMeta(sessionRow, meta)
	riskFlags := s.sessionRiskFlags(ctx, user.ID, sessionRow.ID, meta)
	sessionRow.Status = authSessionStatusActive
	sessionRow.IsSuspicious = len(riskFlags) > 0
	sessionRow.RiskFlags = marshalStringSlice(riskFlags)
	sessionRow.RevokedAt = nil
	sessionRow.RevokedReason = ""
	sessionRow.LastSeenAt = &now
	sessionRow.LastRefreshedAt = &now
	sessionRow.UpdatedAt = now
	accessSession := toSessionUserInfo(user, effective)
	accessSession.TokenID = buildTokenID("access", sessionRow.ID)
	accessToken, err := jwtoken.New(
		jwtCfg.Secret,
		jwtoken.WithIssuer(jwtCfg.Issuer),
		jwtoken.WithAudience(jwtCfg.Audience),
		jwtoken.WithLeeway(time.Duration(jwtCfg.LeewaySeconds)*time.Second),
	).Sign(accessSession, accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshSession := toSessionUserInfo(user, effective)
	refreshSession.TokenID = buildTokenID("refresh", sessionRow.ID)
	refreshToken, err := jwtoken.New(
		jwtCfg.Secret,
		jwtoken.WithIssuer(jwtCfg.Issuer),
		jwtoken.WithAudience(jwtCfg.Audience),
		jwtoken.WithLeeway(time.Duration(jwtCfg.LeewaySeconds)*time.Second),
	).Sign(refreshSession, refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	sessionRow.AccessTokenID = accessSession.TokenID
	sessionRow.RefreshTokenID = refreshSession.TokenID
	sessionRow.RefreshTokenHash = tokenDigest(refreshToken)
	if err := s.store.saveSession(ctx, sessionRow); err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
		UserInfo:     toCurrentUser(user, effective),
		Session:      ptrSessionInfo(toSessionInfo(sessionRow, true)),
	}, nil
}

func (s *service) parseToken(token string) (*jwtokenClaims, error) {
	jwtCfg := configs.Get().JWT
	claims, err := jwtoken.New(
		jwtCfg.Secret,
		jwtoken.WithIssuer(jwtCfg.Issuer),
		jwtoken.WithAudience(jwtCfg.Audience),
		jwtoken.WithLeeway(time.Duration(jwtCfg.LeewaySeconds)*time.Second),
	).Parse(token)
	if err != nil {
		return nil, err
	}
	return &jwtokenClaims{SessionUserInfo: claims.SessionUserInfo}, nil
}

type jwtokenClaims struct {
	proposal.SessionUserInfo
}

func revokeUserSessionsTx(tx *gorm.DB, userID int32, keepSessionID, reason string) error {
	updates := map[string]interface{}{
		"status":         authSessionStatusRevoked,
		"revoked_at":     time.Now(),
		"revoked_reason": strings.TrimSpace(reason),
	}
	query := tx.Model(&authSession{}).Where("user_id = ? AND status = ?", userID, authSessionStatusActive)
	if strings.TrimSpace(keepSessionID) != "" {
		query = query.Where("id <> ?", keepSessionID)
	}
	return query.Updates(updates).Error
}

func extractRawAccessToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	raw, _ := ctx.Value(rawAccessTokenContextKey{}).(string)
	return strings.TrimSpace(raw)
}

type rawAccessTokenContextKey struct{}

type RawAccessTokenContextKey = rawAccessTokenContextKey

func (s *service) sessionRiskFlags(ctx context.Context, userID int32, currentSessionID string, meta AuthClientMeta) []string {
	sessions, err := s.store.listActiveSessions(ctx, userID)
	if err != nil {
		return nil
	}
	flags := make([]string, 0, 3)
	if len(sessions) >= 3 {
		flags = append(flags, "too_many_active_sessions")
	}
	deviceID := buildDeviceID(meta)
	clientIP := sanitizeClientIP(meta.ClientIP)
	for _, item := range sessions {
		if item.ID == currentSessionID {
			continue
		}
		if deviceID != "" && item.DeviceID != "" && item.DeviceID != deviceID {
			flags = appendIfMissing(flags, "new_device")
		}
		if clientIP != "" && item.ClientIP != "" && item.ClientIP != clientIP {
			flags = appendIfMissing(flags, "new_network")
		}
		if item.IsSuspicious {
			flags = appendIfMissing(flags, "risk_chain_detected")
		}
	}
	if clientIP == "" || sanitizeUserAgent(meta.UserAgent) == "" {
		flags = appendIfMissing(flags, "missing_client_fingerprint")
	}
	return flags
}

func (s *service) activeSessionCount(ctx context.Context, userID int32) int {
	return s.store.countActiveSessions(ctx, userID)
}

func (s *service) recoveryRiskFlags(ctx context.Context, userID int32, meta AuthClientMeta) []string {
	flags := make([]string, 0, 3)
	if s.activeSessionCount(ctx, userID) >= 3 {
		flags = append(flags, "too_many_active_sessions")
	}
	if sanitizeClientIP(meta.ClientIP) == "" {
		flags = appendIfMissing(flags, "missing_client_ip")
	}
	if s.store.countSuspiciousSessions(ctx, userID) > 0 {
		flags = appendIfMissing(flags, "suspicious_sessions_present")
	}
	return flags
}

func (s *service) sendRecoveryCode(ctx context.Context, challenge *authRecoveryChallenge, user *sysUser, channel, target, code string) error {
	if s == nil || s.recoveryDelivery == nil {
		return apperr.DependencyFailed("recovery delivery sender is unavailable")
	}
	result, err := s.recoveryDelivery.Send(ctx, RecoveryDeliveryMessage{
		ChallengeID:    challenge.ID,
		UserID:         user.ID,
		Username:       user.Username,
		OrganizationID: derefInt32(user.OrganizationID),
		Channel:        channel,
		Target:         target,
		Code:           code,
		Template:       "password_recovery",
	})
	if err != nil {
		_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.delivery_failed", "error", map[string]any{
			"channel":      channel,
			"targetMasked": maskRecoveryTarget(target, channel),
			"reason":       err.Error(),
			"errorKind":    notification.ErrorKind(err),
			"scene":        notification.ScenePasswordRecovery,
		}, AuthClientMeta{})
		return apperr.DependencyFailed("recovery delivery failed")
	}
	if result == nil || !result.Accepted {
		_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.delivery_failed", "rejected", map[string]any{
			"channel":      channel,
			"targetMasked": maskRecoveryTarget(target, channel),
			"reason":       "recovery delivery rejected",
			"errorKind":    notification.ErrorProviderRejected,
			"scene":        notification.ScenePasswordRecovery,
		}, AuthClientMeta{})
		return apperr.DependencyFailed("recovery delivery rejected")
	}
	maskedTarget := strings.TrimSpace(result.TargetMasked)
	if maskedTarget == "" {
		maskedTarget = maskRecoveryTarget(target, channel)
	}
	_ = s.store.appendRecoveryAudit(ctx, challenge.ID, challenge.UserID, "recovery.delivery_sent", "accepted", map[string]any{
		"channel":      channel,
		"provider":     result.Provider,
		"externalId":   result.ExternalID,
		"targetMasked": maskedTarget,
	}, AuthClientMeta{})
	return nil
}

func (s *service) expireRecoveryChallenge(ctx context.Context, challengeID string) error {
	return s.store.expireRecoveryChallenge(ctx, challengeID)
}

func (s *service) insertRecoveryAudit(ctx context.Context, challengeID string, userID int32, action, status string, detail map[string]any, meta AuthClientMeta) error {
	return s.store.appendRecoveryAudit(ctx, challengeID, userID, action, status, detail, meta)
}

func (s *service) resolveEffectiveProfile(ctx context.Context, user *sysUser) (*tenantservice.EffectiveProfile, error) {
	if user == nil {
		return &tenantservice.EffectiveProfile{SubscriptionStatus: proposal.SubscriptionStatusActive}, nil
	}
	modelUser := &model.SysUser{
		ID:             user.ID,
		RoleCode:       user.RoleCode,
		OrganizationID: user.OrganizationID,
		CampusID:       user.CampusID,
	}
	return s.tenantService.ResolveEffectiveProfile(ctx, modelUser)
}

func validateEffectiveProfile(user *sysUser, effective *tenantservice.EffectiveProfile) error {
	if user == nil {
		return gorm.ErrRecordNotFound
	}
	if effective == nil {
		return nil
	}
	if effective.OrganizationStatus != "" && effective.OrganizationStatus != proposal.OrganizationStatusActive {
		return apperr.Forbidden("所属机构不可用")
	}
	if effective.CampusStatus != "" && effective.CampusStatus != proposal.OrganizationStatusActive {
		return apperr.Forbidden("所属校区不可用")
	}
	if !proposal.IsActiveSubscriptionStatus(effective.SubscriptionStatus) {
		return apperr.Forbidden("机构订阅不可用")
	}
	if len(effective.FeatureFlags) > 0 {
		featureRequired := proposal.FeatureAuth
		found := false
		for _, item := range effective.FeatureFlags {
			if item == featureRequired {
				found = true
				break
			}
		}
		if !found {
			return apperr.Forbidden("当前租户未开通认证功能")
		}
	}
	return nil
}

func nullableInt32(v int32) *int32 {
	if v <= 0 {
		return nil
	}
	return &v
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func insertRecoveryAuditTx(tx *gorm.DB, challengeID string, userID int32, action, status string, detail map[string]any, meta AuthClientMeta) error {
	detailRaw := ""
	if len(detail) > 0 {
		buf, _ := json.Marshal(detail)
		detailRaw = string(buf)
		if len(detailRaw) > recoveryAuditDetailMaxBytes {
			detailRaw = detailRaw[:recoveryAuditDetailMaxBytes]
		}
	}
	row := &authRecoveryAudit{
		ID:          idgen.GenerateUniqueID(),
		ChallengeID: strings.TrimSpace(challengeID),
		UserID:      userID,
		Action:      strings.TrimSpace(action),
		Status:      strings.TrimSpace(status),
		Detail:      detailRaw,
		ActorIP:     sanitizeClientIP(meta.ClientIP),
		UserAgent:   sanitizeUserAgent(meta.UserAgent),
		CreatedAt:   time.Now(),
	}
	return tx.Create(row).Error
}

func ensureUserAvailable(user *sysUser) error {
	if user == nil {
		return gorm.ErrRecordNotFound
	}
	if user.Status == proposal.UserStatusLocked {
		return ensureUserUnlocked(user)
	}
	if user.Status == proposal.UserStatusDisabled {
		return apperr.Forbidden("账号已禁用")
	}
	return nil
}

func ensureUserUnlocked(user *sysUser) error {
	if user == nil {
		return gorm.ErrRecordNotFound
	}
	if user.LockedUntil == nil {
		if user.Status == proposal.UserStatusLocked {
			return apperr.Forbidden("账号已锁定，请联系管理员")
		}
		return nil
	}
	if time.Now().Before(*user.LockedUntil) {
		return apperr.Forbidden(fmt.Sprintf("账号已锁定，请于 %s 后重试", user.LockedUntil.Local().Format("2006-01-02 15:04:05")))
	}
	return nil
}

func toCurrentUser(user *sysUser, effective *tenantservice.EffectiveProfile) *CurrentUser {
	if user == nil {
		return nil
	}
	profile := proposal.BuildAccessProfile(user.RoleCode)
	current := &CurrentUser{
		ID:                 user.ID,
		Username:           user.Username,
		RealName:           user.RealName,
		RoleCode:           user.RoleCode,
		Phone:              user.Phone,
		Email:              user.Email,
		Status:             user.Status,
		MustChangePassword: user.MustChangePassword,
		Remark:             user.Remark,
		Permissions:        append([]string(nil), profile.Permissions...),
		MenuPermissions:    append([]string(nil), profile.MenuPermissions...),
		DataScope:          profile.DataScope,
	}
	if effective != nil {
		current.OrganizationID = effective.OrganizationID
		current.CampusID = effective.CampusID
		current.OrganizationName = effective.OrganizationName
		current.CampusName = effective.CampusName
		current.SubscriptionStatus = effective.SubscriptionStatus
		current.FeatureFlags = append([]string(nil), effective.FeatureFlags...)
	}
	return current
}

func toSessionUserInfo(user *sysUser, effective *tenantservice.EffectiveProfile) proposal.SessionUserInfo {
	if user == nil {
		return proposal.SessionUserInfo{}
	}
	profile := proposal.BuildAccessProfile(user.RoleCode)
	session := proposal.SessionUserInfo{
		Id:                 user.ID,
		UserName:           user.Username,
		NickName:           user.RealName,
		RoleCode:           user.RoleCode,
		Status:             user.Status,
		MustChangePassword: user.MustChangePassword,
		Permissions:        append([]string(nil), profile.Permissions...),
		MenuPermissions:    append([]string(nil), profile.MenuPermissions...),
		DataScope:          profile.DataScope,
	}
	if effective != nil {
		session.OrganizationID = effective.OrganizationID
		session.CampusID = effective.CampusID
		session.OrganizationName = effective.OrganizationName
		session.CampusName = effective.CampusName
		session.SubscriptionStatus = effective.SubscriptionStatus
		session.FeatureFlags = append([]string(nil), effective.FeatureFlags...)
	}
	return session
}

func toSessionInfo(row *authSession, current bool) SessionInfo {
	if row == nil {
		return SessionInfo{}
	}
	return SessionInfo{
		ID:              row.ID,
		DeviceID:        row.DeviceID,
		DeviceName:      row.DeviceName,
		ClientIP:        row.ClientIP,
		UserAgent:       row.UserAgent,
		Status:          row.Status,
		Current:         current,
		Suspicious:      row.IsSuspicious,
		OrganizationID:  derefInt32(row.OrganizationID),
		CampusID:        derefInt32(row.CampusID),
		RoleCode:        row.RoleCode,
		RiskFlags:       unmarshalStringSlice(row.RiskFlags),
		LastSeenAt:      row.LastSeenAt,
		LastRefreshedAt: row.LastRefreshedAt,
		RevokedAt:       row.RevokedAt,
		RevokedReason:   row.RevokedReason,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func ptrSessionInfo(v SessionInfo) *SessionInfo { return &v }

func maxFailedLoginAttempts() int32 { return defaultMaxFailedAttempts }
func lockDuration() time.Duration   { return defaultTemporaryLockTTL }

func buildTokenID(kind, sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", strings.TrimSpace(kind), strings.TrimSpace(sessionID), idgen.GenerateUniqueID())
}

func ExtractSessionID(tokenID string) string {
	_, sessionID := ParseTokenKindAndSessionID(tokenID)
	return sessionID
}

func ParseTokenKindAndSessionID(tokenID string) (string, string) {
	parts := strings.Split(strings.TrimSpace(tokenID), ":")
	if len(parts) < 3 {
		return "", ""
	}
	return parts[0], parts[1]
}

func tokenDigest(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func applyClientMeta(row *authSession, meta AuthClientMeta) {
	if row == nil {
		return
	}
	deviceName := normalizeDeviceName(meta)
	if deviceName != "" {
		row.DeviceName = deviceName
	}
	if ip := sanitizeClientIP(meta.ClientIP); ip != "" {
		row.ClientIP = ip
	}
	if ua := sanitizeUserAgent(meta.UserAgent); ua != "" {
		row.UserAgent = ua
		row.UserAgentHash = tokenDigest(ua)
	}
	if row.DeviceID == "" || meta.DeviceName != "" || meta.UserAgent != "" {
		row.DeviceID = buildDeviceID(meta)
	}
}

func buildDeviceID(meta AuthClientMeta) string {
	fingerprint := strings.TrimSpace(normalizeDeviceName(meta) + "|" + sanitizeUserAgent(meta.UserAgent))
	if fingerprint == "" {
		return ""
	}
	return tokenDigest(fingerprint)[:32]
}

func normalizeDeviceName(meta AuthClientMeta) string {
	deviceName := strings.TrimSpace(meta.DeviceName)
	if deviceName != "" {
		return truncateString(deviceName, 120)
	}
	ua := sanitizeUserAgent(meta.UserAgent)
	if ua == "" {
		return "unknown-device"
	}
	return truncateString(ua, 120)
}

func sanitizeClientIP(ip string) string {
	return truncateString(strings.TrimSpace(ip), 64)
}

func sanitizeUserAgent(ua string) string {
	return truncateString(strings.TrimSpace(ua), 255)
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

func marshalStringSlice(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	buf, _ := json.Marshal(values)
	return string(buf)
}

func unmarshalStringSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	return result
}

func appendIfMissing(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}

func normalizeRecoveryChannel(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "sms":
		return "sms"
	case "email", "mail":
		return "email"
	default:
		return ""
	}
}

func recoveryTarget(user *sysUser, channel string) string {
	if user == nil {
		return ""
	}
	switch channel {
	case "sms":
		return strings.TrimSpace(user.Phone)
	case "email":
		return strings.TrimSpace(user.Email)
	default:
		return ""
	}
}

func shouldRequireSecondFactor(user *sysUser, meta AuthClientMeta, activeSessions int) bool {
	if user == nil {
		return false
	}
	if user.RoleCode == proposal.RolePlatformAdmin {
		return true
	}
	if activeSessions >= 3 {
		return true
	}
	return sanitizeClientIP(meta.ClientIP) == ""
}

func defaultSecondFactorChannel(user *sysUser, primary string) string {
	if user == nil {
		return ""
	}
	if primary != "sms" && strings.TrimSpace(user.Phone) != "" {
		return "sms"
	}
	if primary != "email" && strings.TrimSpace(user.Email) != "" {
		return "email"
	}
	return ""
}

func maskRecoveryTarget(target, channel string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	switch channel {
	case "sms":
		if len(target) <= 4 {
			return "****"
		}
		return target[:3] + strings.Repeat("*", maxInt(1, len(target)-7)) + target[len(target)-4:]
	case "email":
		parts := strings.Split(target, "@")
		if len(parts) != 2 {
			return "***"
		}
		local := parts[0]
		if len(local) <= 2 {
			local = local[:1] + "***"
		} else {
			local = local[:2] + "***"
		}
		return local + "@" + parts[1]
	default:
		return "***"
	}
}

func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}
	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}
	return builder.String(), nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
