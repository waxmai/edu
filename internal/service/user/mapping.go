package user

func toUserItem(item sysUser) UserItem {
	return UserItem{
		ID:                    item.ID,
		Username:              item.Username,
		RoleCode:              item.RoleCode,
		OrganizationID:        derefInt32(item.OrganizationID),
		CampusID:              derefInt32(item.CampusID),
		RealName:              item.RealName,
		Phone:                 item.Phone,
		Email:                 item.Email,
		Status:                item.Status,
		FailedLoginCount:      item.FailedLoginCount,
		LockedUntil:           item.LockedUntil,
		MustChangePassword:    item.MustChangePassword,
		LastPasswordChangedAt: item.LastPasswordChangedAt,
		Remark:                item.Remark,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
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
