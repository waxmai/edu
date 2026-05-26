package interceptor

import (
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	serviceAuth "edu-schedule-system/internal/service/auth"

	"go.uber.org/zap"
)

type Interceptor interface {
	// Authenticate 示例验证用户身份
	Authenticate() core.HandlerFunc

	// RequirePasswordChanged 限制首次登录未改密用户访问业务接口
	RequirePasswordChanged() core.HandlerFunc

	// RequireTenantActive 限制租户停用或过期后的关键业务访问
	RequireTenantActive() core.HandlerFunc

	// RequireRoles 角色校验
	RequireRoles(roles ...string) core.HandlerFunc

	// RequireAdmin 管理员校验
	RequireAdmin() core.HandlerFunc

	// RequirePermissions 权限点校验
	RequirePermissions(permissions ...string) core.HandlerFunc
	// RequirePlatformRoles 平台预置角色校验
	RequirePlatformRoles() core.HandlerFunc

	// RequireFeature 功能开关校验
	RequireFeature(feature string) core.HandlerFunc

	// RequireDataScopeSelfOrAll 数据范围校验
	RequireDataScopeSelfOrAll() core.HandlerFunc

	// JWTokenAuthVerify JWT token 授权验证
	JWTokenAuthVerify(ctx core.Context) (sessionUserInfo proposal.SessionUserInfo, err core.BusinessError)

	// i 为了避免被其他包实现
	i()
}

type interceptor struct {
	logger       *zap.Logger
	db           mysql.Repo
	tokenRevoker serviceAuth.TokenRevoker
}

func New(logger *zap.Logger, db mysql.Repo, opts ...Option) Interceptor {
	it := &interceptor{
		logger:       logger,
		db:           db,
		tokenRevoker: serviceAuth.NewTokenRevoker(nil),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(it)
		}
	}
	return it
}

func (i *interceptor) i() {}

type Option func(*interceptor)

func WithTokenRevoker(revoker serviceAuth.TokenRevoker) Option {
	return func(i *interceptor) {
		if i == nil || revoker == nil {
			return
		}
		i.tokenRevoker = revoker
	}
}
