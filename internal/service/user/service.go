package user

import (
	"context"
	"time"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql"
	tenantservice "edu-schedule-system/internal/service/tenant"
)

type Service interface {
	List(ctx context.Context, actor proposal.SessionUserInfo, query ListQuery) (*ListResponse, error)
	Create(ctx context.Context, actor proposal.SessionUserInfo, req *CreateRequest) (int32, error)
	Update(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateRequest) error
	UpdateStatus(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateStatusRequest) error
	ResetPassword(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *ResetPasswordRequest) error
	Unlock(ctx context.Context, actor proposal.SessionUserInfo, id int32) error
}

type service struct {
	db            mysql.Repo
	tenantService tenantservice.Service
}

type sysUser struct {
	ID                    int32      `gorm:"column:id;primaryKey" json:"id"`
	Username              string     `gorm:"column:username" json:"username"`
	PasswordHash          string     `gorm:"column:password_hash" json:"-"`
	RoleCode              string     `gorm:"column:role_code" json:"roleCode"`
	OrganizationID        *int32     `gorm:"column:organization_id" json:"organizationId"`
	CampusID              *int32     `gorm:"column:campus_id" json:"campusId"`
	RealName              string     `gorm:"column:real_name" json:"realName"`
	Phone                 string     `gorm:"column:phone" json:"phone"`
	Email                 string     `gorm:"column:email" json:"email"`
	Status                string     `gorm:"column:status" json:"status"`
	FailedLoginCount      int32      `gorm:"column:failed_login_count" json:"failedLoginCount"`
	LockedUntil           *time.Time `gorm:"column:locked_until" json:"lockedUntil"`
	MustChangePassword    bool       `gorm:"column:must_change_password" json:"mustChangePassword"`
	LastPasswordChangedAt *time.Time `gorm:"column:last_password_changed_at" json:"lastPasswordChangedAt"`
	Remark                string     `gorm:"column:remark" json:"remark"`
	CreatedAt             time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt             time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (sysUser) TableName() string { return "sys_user" }

func New(db mysql.Repo) Service {
	return &service{db: db, tenantService: tenantservice.New(db)}
}

type ListQuery struct {
	Username       string `form:"username"`
	RealName       string `form:"realName"`
	RoleCode       string `form:"roleCode"`
	OrganizationID int32  `form:"organizationId"`
	CampusID       int32  `form:"campusId"`
	Status         string `form:"status"`
	PageNum        int    `form:"pageNum"`
	PageSize       int    `form:"pageSize"`
}

type UserItem struct {
	ID                    int32      `json:"id"`
	Username              string     `json:"username"`
	RoleCode              string     `json:"roleCode"`
	OrganizationID        int32      `json:"organizationId"`
	CampusID              int32      `json:"campusId"`
	RealName              string     `json:"realName"`
	Phone                 string     `json:"phone"`
	Email                 string     `json:"email"`
	Status                string     `json:"status"`
	FailedLoginCount      int32      `json:"failedLoginCount"`
	LockedUntil           *time.Time `json:"lockedUntil"`
	MustChangePassword    bool       `json:"mustChangePassword"`
	LastPasswordChangedAt *time.Time `json:"lastPasswordChangedAt"`
	Remark                string     `json:"remark"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type ListResponse struct {
	List  []UserItem `json:"list"`
	Total int64      `json:"total"`
}

type CreateRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password"`
	RoleCode       string `json:"roleCode" binding:"required"`
	OrganizationID int32  `json:"organizationId"`
	CampusID       int32  `json:"campusId"`
	RealName       string `json:"realName" binding:"required"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

type UpdateRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password"`
	RoleCode       string `json:"roleCode" binding:"required"`
	OrganizationID int32  `json:"organizationId"`
	CampusID       int32  `json:"campusId"`
	RealName       string `json:"realName" binding:"required"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword"`
}
