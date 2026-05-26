package user

import (
	"fmt"
	"strings"
	"unicode"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
)

const minPasswordLength = 12

func requireUserWrite(actor proposal.SessionUserInfo) error {
	if actor.IsPlatformAdmin() || actor.IsOrgAdmin() || actor.IsCampusAdmin() {
		return nil
	}
	return apperr.Forbidden("admin permission is required")
}

func validateCreateRequest(req *CreateRequest) (*CreateRequest, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("user create request is required")
	}
	username := strings.TrimSpace(req.Username)
	realName := strings.TrimSpace(req.RealName)
	roleCode := strings.TrimSpace(req.RoleCode)
	password := strings.TrimSpace(req.Password)
	if username == "" || realName == "" || roleCode == "" {
		return nil, apperr.InvalidArgument("username, realName, roleCode are required")
	}
	if !isValidRole(roleCode) {
		return nil, apperr.InvalidArgument("roleCode must be platform_admin, org_admin, campus_admin, or teacher")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = proposal.UserStatusEnabled
	}
	if !isValidStatus(status) {
		return nil, apperr.InvalidArgument("status must be enabled, disabled, or locked")
	}
	if password == "" {
		return nil, apperr.InvalidArgument("password is required")
	}
	if err := validatePasswordStrength(password); err != nil {
		return nil, err
	}
	return &CreateRequest{
		Username:       username,
		Password:       password,
		RoleCode:       roleCode,
		OrganizationID: req.OrganizationID,
		CampusID:       req.CampusID,
		RealName:       realName,
		Phone:          strings.TrimSpace(req.Phone),
		Email:          strings.TrimSpace(req.Email),
		Status:         status,
		Remark:         strings.TrimSpace(req.Remark),
	}, nil
}

func validateUpdateRequest(req *UpdateRequest) (*UpdateRequest, error) {
	if req == nil {
		return nil, apperr.InvalidArgument("user update request is required")
	}
	username := strings.TrimSpace(req.Username)
	realName := strings.TrimSpace(req.RealName)
	roleCode := strings.TrimSpace(req.RoleCode)
	if username == "" || realName == "" || roleCode == "" {
		return nil, apperr.InvalidArgument("username, realName, roleCode are required")
	}
	if !isValidRole(roleCode) {
		return nil, apperr.InvalidArgument("roleCode must be platform_admin, org_admin, campus_admin, or teacher")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = proposal.UserStatusEnabled
	}
	if !isValidStatus(status) {
		return nil, apperr.InvalidArgument("status must be enabled, disabled, or locked")
	}
	password := strings.TrimSpace(req.Password)
	if password != "" {
		if err := validatePasswordStrength(password); err != nil {
			return nil, err
		}
	}
	return &UpdateRequest{
		Username:       username,
		Password:       password,
		RoleCode:       roleCode,
		OrganizationID: req.OrganizationID,
		CampusID:       req.CampusID,
		RealName:       realName,
		Phone:          strings.TrimSpace(req.Phone),
		Email:          strings.TrimSpace(req.Email),
		Status:         status,
		Remark:         strings.TrimSpace(req.Remark),
	}, nil
}

func validatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return apperr.InvalidArgument(fmt.Sprintf("password length must be at least %d characters", minPasswordLength))
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return apperr.InvalidArgument("password must include upper-case, lower-case, digit, and special character")
	}
	return nil
}

func isValidRole(role string) bool {
	return role == proposal.RolePlatformAdmin || role == proposal.RoleOrgAdmin || role == proposal.RoleCampusAdmin || role == proposal.RoleTeacher
}

func isValidStatus(status string) bool {
	return status == proposal.UserStatusEnabled || status == proposal.UserStatusDisabled || status == proposal.UserStatusLocked
}
