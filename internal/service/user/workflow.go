package user

import (
	"context"
	"strings"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/serviceutil"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *service) List(ctx context.Context, actor proposal.SessionUserInfo, query ListQuery) (*ListResponse, error) {
	if err := requireUserWrite(actor); err != nil {
		return nil, err
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	if query.OrganizationID > 0 {
		if err := serviceutil.EnsureSameOrganization(actor, query.OrganizationID); err != nil {
			return nil, err
		}
	}
	if query.CampusID > 0 {
		if err := serviceutil.EnsureSameCampus(actor, query.CampusID); err != nil {
			return nil, err
		}
	}

	dbq := s.db.GetDbR().WithContext(ctx).Model(&sysUser{})
	if !actor.IsPlatformAdmin() {
		if actor.OrganizationID > 0 {
			dbq = dbq.Where("organization_id = ?", actor.OrganizationID)
		}
		if actor.IsCampusAdmin() && actor.CampusID > 0 {
			dbq = dbq.Where("campus_id = ?", actor.CampusID)
		}
	}
	if v := strings.TrimSpace(query.Username); v != "" {
		dbq = dbq.Where("username LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(query.RealName); v != "" {
		dbq = dbq.Where("real_name LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(query.RoleCode); v != "" {
		dbq = dbq.Where("role_code = ?", v)
	}
	if query.OrganizationID > 0 {
		dbq = dbq.Where("organization_id = ?", query.OrganizationID)
	}
	if query.CampusID > 0 {
		dbq = dbq.Where("campus_id = ?", query.CampusID)
	}
	if v := strings.TrimSpace(query.Status); v != "" {
		dbq = dbq.Where("status = ?", v)
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []sysUser
	if err := dbq.Order("id desc").Offset((query.PageNum - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	result := &ListResponse{List: make([]UserItem, 0, len(items)), Total: total}
	for _, item := range items {
		result.List = append(result.List, toUserItem(item))
	}
	return result, nil
}

func (s *service) Create(ctx context.Context, actor proposal.SessionUserInfo, req *CreateRequest) (int32, error) {
	if err := requireUserWrite(actor); err != nil {
		return 0, err
	}
	payload, err := validateCreateRequest(req)
	if err != nil {
		return 0, err
	}
	orgID, campusID, err := normalizeManagedScope(actor, payload.RoleCode, payload.OrganizationID, payload.CampusID)
	if err != nil {
		return 0, err
	}
	if err := ensureActorCanAssignScope(actor, payload.RoleCode, orgID, campusID); err != nil {
		return 0, err
	}
	if err := s.ensureUserQuota(ctx, actor, orgID); err != nil {
		return 0, err
	}

	var count int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&sysUser{}).Where("username = ?", payload.Username).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, apperr.Conflict("用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	item := &sysUser{
		Username:           payload.Username,
		PasswordHash:       string(hash),
		RoleCode:           payload.RoleCode,
		OrganizationID:     nullableInt32(orgID),
		CampusID:           nullableInt32(campusID),
		RealName:           payload.RealName,
		Phone:              payload.Phone,
		Email:              payload.Email,
		Status:             payload.Status,
		MustChangePassword: true,
		Remark:             payload.Remark,
	}
	if err := s.db.GetDbW().WithContext(ctx).Create(item).Error; err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *service) Update(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateRequest) error {
	if err := requireUserWrite(actor); err != nil {
		return err
	}
	if id <= 0 {
		return apperr.InvalidArgument("user id must be positive")
	}
	payload, err := validateUpdateRequest(req)
	if err != nil {
		return err
	}

	var item sysUser
	if err := s.db.GetDbW().WithContext(ctx).First(&item, id).Error; err != nil {
		return err
	}
	if err := ensureActorCanManageExistingUser(actor, &item); err != nil {
		return err
	}

	var count int64
	if err := s.db.GetDbR().WithContext(ctx).Model(&sysUser{}).Where("username = ? AND id <> ?", payload.Username, id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return apperr.Conflict("用户名已存在")
	}
	orgID, campusID, err := normalizeManagedScope(actor, payload.RoleCode, payload.OrganizationID, payload.CampusID)
	if err != nil {
		return err
	}
	if err := ensureActorCanAssignScope(actor, payload.RoleCode, orgID, campusID); err != nil {
		return err
	}

	item.Username = payload.Username
	item.RoleCode = payload.RoleCode
	item.OrganizationID = nullableInt32(orgID)
	item.CampusID = nullableInt32(campusID)
	item.RealName = payload.RealName
	item.Phone = payload.Phone
	item.Email = payload.Email
	item.Status = payload.Status
	item.Remark = payload.Remark
	if payload.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		item.PasswordHash = string(hash)
		item.MustChangePassword = true
		item.LastPasswordChangedAt = nil
		item.FailedLoginCount = 0
		item.LockedUntil = nil
	}

	return s.db.GetDbW().WithContext(ctx).Save(&item).Error
}

func (s *service) UpdateStatus(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *UpdateStatusRequest) error {
	if err := requireUserWrite(actor); err != nil {
		return err
	}
	if id <= 0 {
		return apperr.InvalidArgument("user id must be positive")
	}
	status := strings.TrimSpace(req.Status)
	if !isValidStatus(status) {
		return apperr.InvalidArgument("status must be enabled, disabled, or locked")
	}

	var existing sysUser
	if err := s.db.GetDbW().WithContext(ctx).Where("id = ?", id).First(&existing).Error; err != nil {
		return err
	}
	if err := ensureActorCanManageExistingUser(actor, &existing); err != nil {
		return err
	}
	if existing.Status == status && !(status != proposal.UserStatusLocked && existing.LockedUntil != nil) {
		return nil
	}

	updates := map[string]interface{}{"status": status}
	if status != proposal.UserStatusLocked {
		updates["locked_until"] = nil
	}
	result := s.db.GetDbW().WithContext(ctx).Model(&sysUser{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *service) ResetPassword(ctx context.Context, actor proposal.SessionUserInfo, id int32, req *ResetPasswordRequest) error {
	if err := requireUserWrite(actor); err != nil {
		return err
	}
	if id <= 0 {
		return apperr.InvalidArgument("user id must be positive")
	}
	var existing sysUser
	if err := s.db.GetDbW().WithContext(ctx).Where("id = ?", id).First(&existing).Error; err != nil {
		return err
	}
	if err := ensureActorCanManageExistingUser(actor, &existing); err != nil {
		return err
	}
	if req == nil || strings.TrimSpace(req.NewPassword) == "" {
		return apperr.InvalidArgument("newPassword is required")
	}
	password := strings.TrimSpace(req.NewPassword)
	if err := validatePasswordStrength(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	result := s.db.GetDbW().WithContext(ctx).Model(&sysUser{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash":            string(hash),
		"must_change_password":     true,
		"last_password_changed_at": nil,
		"failed_login_count":       0,
		"locked_until":             nil,
		"status":                   proposal.UserStatusEnabled,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *service) Unlock(ctx context.Context, actor proposal.SessionUserInfo, id int32) error {
	if err := requireUserWrite(actor); err != nil {
		return err
	}
	if id <= 0 {
		return apperr.InvalidArgument("user id must be positive")
	}
	var existing sysUser
	if err := s.db.GetDbW().WithContext(ctx).Where("id = ?", id).First(&existing).Error; err != nil {
		return err
	}
	if err := ensureActorCanManageExistingUser(actor, &existing); err != nil {
		return err
	}
	result := s.db.GetDbW().WithContext(ctx).Model(&sysUser{}).Where("id = ?", id).Updates(map[string]interface{}{
		"failed_login_count": 0,
		"locked_until":       nil,
		"status":             proposal.UserStatusEnabled,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
