package proposal

import "encoding/json"

const (
	RolePlatformAdmin   = "platform_admin"
	RolePlatformOps     = "platform_ops"
	RolePlatformFinance = "platform_finance"
	RolePlatformAuditor = "platform_auditor"
	RolePlatformSupport = "platform_support"
	RoleOrgAdmin        = "org_admin"
	RoleCampusAdmin     = "campus_admin"
	RoleTeacher         = "teacher"

	UserStatusEnabled  = "enabled"
	UserStatusDisabled = "disabled"
	UserStatusLocked   = "locked"
)

// SessionUserInfo 当前用户会话信息
type SessionUserInfo struct {
	Id                 int32    `json:"id"`                   // ID
	UserName           string   `json:"username"`             // 用户名
	NickName           string   `json:"nickname"`             // 昵称
	RoleCode           string   `json:"role_code"`            // 角色编码
	Status             string   `json:"status"`               // 用户状态
	OrganizationID     int32    `json:"organization_id"`      // 所属机构ID
	CampusID           int32    `json:"campus_id"`            // 所属校区ID
	OrganizationName   string   `json:"organization_name"`    // 所属机构名称
	CampusName         string   `json:"campus_name"`          // 所属校区名称
	SubscriptionStatus string   `json:"subscription_status"`  // 订阅状态
	FeatureFlags       []string `json:"feature_flags"`        // 生效功能开关
	MustChangePassword bool     `json:"must_change_password"` // 是否必须修改密码
	TokenID            string   `json:"token_id"`             // 当前 token 唯一标识
	Permissions        []string `json:"permissions"`          // 权限点
	MenuPermissions    []string `json:"menu_permissions"`     // 菜单权限点
	DataScope          string   `json:"data_scope"`           // 数据范围
}

func (user SessionUserInfo) IsPlatformAdmin() bool {
	return user.RoleCode == RolePlatformAdmin
}

func (user SessionUserInfo) IsPlatformRole() bool {
	switch user.RoleCode {
	case RolePlatformAdmin, RolePlatformOps, RolePlatformFinance, RolePlatformAuditor, RolePlatformSupport:
		return true
	default:
		return false
	}
}

func (user SessionUserInfo) IsOrgAdmin() bool {
	return user.RoleCode == RoleOrgAdmin
}

func (user SessionUserInfo) IsCampusAdmin() bool {
	return user.RoleCode == RoleCampusAdmin
}

func (user SessionUserInfo) IsAdmin() bool {
	return user.IsPlatformAdmin() || user.IsOrgAdmin() || user.IsCampusAdmin()
}

func (user SessionUserInfo) IsTeacher() bool {
	return user.RoleCode == RoleTeacher
}

func (user SessionUserInfo) IsEnabled() bool {
	return user.Status == "" || user.Status == UserStatusEnabled
}

func (user SessionUserInfo) HasFeature(feature string) bool {
	if feature == "" {
		return true
	}
	for _, item := range user.FeatureFlags {
		if item == feature {
			return true
		}
	}
	return false
}

// Marshal 序列化到JSON
func (user *SessionUserInfo) Marshal() (jsonRaw []byte) {
	jsonRaw, _ = json.Marshal(user)
	return
}

func (user SessionUserInfo) HasPermission(permission string) bool {
	for _, item := range user.Permissions {
		if item == permission {
			return true
		}
	}
	return false
}

func (user SessionUserInfo) HasAnyPermission(permissions ...string) bool {
	if len(permissions) == 0 {
		return true
	}
	for _, permission := range permissions {
		if user.HasPermission(permission) {
			return true
		}
	}
	return false
}
