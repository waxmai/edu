package auth

import "time"

type sysUser struct {
	ID                    int32      `gorm:"column:id;primaryKey"`
	Username              string     `gorm:"column:username"`
	PasswordHash          string     `gorm:"column:password_hash"`
	RoleCode              string     `gorm:"column:role_code"`
	OrganizationID        *int32     `gorm:"column:organization_id"`
	CampusID              *int32     `gorm:"column:campus_id"`
	RealName              string     `gorm:"column:real_name"`
	Phone                 string     `gorm:"column:phone"`
	Email                 string     `gorm:"column:email"`
	Status                string     `gorm:"column:status"`
	FailedLoginCount      int32      `gorm:"column:failed_login_count"`
	LockedUntil           *time.Time `gorm:"column:locked_until"`
	MustChangePassword    bool       `gorm:"column:must_change_password"`
	LastPasswordChangedAt *time.Time `gorm:"column:last_password_changed_at"`
	Remark                string     `gorm:"column:remark"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (sysUser) TableName() string { return "sys_user" }

type authSession struct {
	ID                 string     `gorm:"column:id;primaryKey"`
	UserID             int32      `gorm:"column:user_id"`
	OrganizationID     *int32     `gorm:"column:organization_id"`
	CampusID           *int32     `gorm:"column:campus_id"`
	RoleCode           string     `gorm:"column:role_code"`
	DeviceID           string     `gorm:"column:device_id"`
	DeviceName         string     `gorm:"column:device_name"`
	ClientIP           string     `gorm:"column:client_ip"`
	UserAgent          string     `gorm:"column:user_agent"`
	UserAgentHash      string     `gorm:"column:user_agent_hash"`
	AccessTokenID      string     `gorm:"column:access_token_id"`
	RefreshTokenID     string     `gorm:"column:refresh_token_id"`
	RefreshTokenHash   string     `gorm:"column:refresh_token_hash"`
	Status             string     `gorm:"column:status"`
	IsSuspicious       bool       `gorm:"column:is_suspicious"`
	RiskFlags          string     `gorm:"column:risk_flags"`
	LastSeenAt         *time.Time `gorm:"column:last_seen_at"`
	LastRefreshedAt    *time.Time `gorm:"column:last_refreshed_at"`
	RevokedAt          *time.Time `gorm:"column:revoked_at"`
	RevokedReason      string     `gorm:"column:revoked_reason"`
	RecoveryVerifiedAt *time.Time `gorm:"column:recovery_verified_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (authSession) TableName() string { return "auth_session" }

type authRecoveryChallenge struct {
	ID                  string     `gorm:"column:id;primaryKey"`
	UserID              int32      `gorm:"column:user_id"`
	Username            string     `gorm:"column:username"`
	Channel             string     `gorm:"column:channel"`
	ChannelTarget       string     `gorm:"column:channel_target"`
	CodeHash            string     `gorm:"column:code_hash"`
	SecondChannel       string     `gorm:"column:second_channel"`
	SecondChannelTarget string     `gorm:"column:second_channel_target"`
	SecondCodeHash      string     `gorm:"column:second_code_hash"`
	Status              string     `gorm:"column:status"`
	RiskFlags           string     `gorm:"column:risk_flags"`
	RequestedIP         string     `gorm:"column:requested_ip"`
	RequestedUserAgent  string     `gorm:"column:requested_user_agent"`
	ExpiresAt           time.Time  `gorm:"column:expires_at"`
	VerifiedAt          *time.Time `gorm:"column:verified_at"`
	ConsumedAt          *time.Time `gorm:"column:consumed_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (authRecoveryChallenge) TableName() string { return "auth_recovery_challenge" }

type authRecoveryAudit struct {
	ID          string    `gorm:"column:id;primaryKey"`
	ChallengeID string    `gorm:"column:challenge_id"`
	UserID      int32     `gorm:"column:user_id"`
	Action      string    `gorm:"column:action"`
	Status      string    `gorm:"column:status"`
	Detail      string    `gorm:"column:detail"`
	ActorIP     string    `gorm:"column:actor_ip"`
	UserAgent   string    `gorm:"column:user_agent"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (authRecoveryAudit) TableName() string { return "auth_recovery_audit" }
