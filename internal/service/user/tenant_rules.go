package user

import (
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
)

func ensureActorCanAssignScope(actor proposal.SessionUserInfo, roleCode string, organizationID, campusID int32) error {
	switch {
	case actor.IsPlatformAdmin():
		switch roleCode {
		case proposal.RolePlatformAdmin:
			if organizationID > 0 || campusID > 0 {
				return apperr.InvalidArgument("platform admin cannot bind organizationId or campusId")
			}
		case proposal.RoleOrgAdmin:
			if organizationID <= 0 {
				return apperr.InvalidArgument("org admin must belong to an organization")
			}
			if campusID > 0 {
				return apperr.InvalidArgument("org admin cannot bind campusId")
			}
		case proposal.RoleCampusAdmin, proposal.RoleTeacher:
			if organizationID <= 0 || campusID <= 0 {
				return apperr.InvalidArgument("campus admin and teacher must belong to organization and campus")
			}
		}
		return nil
	case actor.IsOrgAdmin():
		if roleCode == proposal.RolePlatformAdmin {
			return apperr.Forbidden("org admin cannot manage platform admins")
		}
		if organizationID != actor.OrganizationID {
			return apperr.Forbidden("organization scope mismatch")
		}
		if roleCode == proposal.RoleOrgAdmin && campusID > 0 {
			return apperr.InvalidArgument("org admin cannot bind campusId")
		}
		if (roleCode == proposal.RoleTeacher || roleCode == proposal.RoleCampusAdmin) && campusID <= 0 {
			return apperr.InvalidArgument("teacher and campus admin must belong to a campus")
		}
		return nil
	case actor.IsCampusAdmin():
		if roleCode != proposal.RoleTeacher {
			return apperr.Forbidden("campus admin can only manage teachers")
		}
		if organizationID != actor.OrganizationID || campusID != actor.CampusID {
			return apperr.Forbidden("campus scope mismatch")
		}
		return nil
	default:
		return apperr.Forbidden("admin permission is required")
	}
}

func normalizeManagedScope(actor proposal.SessionUserInfo, roleCode string, organizationID, campusID int32) (int32, int32, error) {
	switch {
	case actor.IsPlatformAdmin():
		if roleCode == proposal.RolePlatformAdmin {
			return 0, 0, nil
		}
		if roleCode == proposal.RoleOrgAdmin {
			if organizationID <= 0 {
				return 0, 0, apperr.InvalidArgument("organizationId is required")
			}
			return organizationID, 0, nil
		}
		if organizationID <= 0 {
			return 0, 0, apperr.InvalidArgument("organizationId is required")
		}
		if roleCode == proposal.RoleTeacher || roleCode == proposal.RoleCampusAdmin {
			if campusID <= 0 {
				return 0, 0, apperr.InvalidArgument("campusId is required")
			}
		}
		return organizationID, campusID, nil
	case actor.IsOrgAdmin():
		orgID := actor.OrganizationID
		if orgID <= 0 {
			return 0, 0, apperr.Forbidden("org admin missing organization scope")
		}
		if roleCode == proposal.RolePlatformAdmin {
			return 0, 0, apperr.Forbidden("org admin cannot manage platform admins")
		}
		if roleCode == proposal.RoleOrgAdmin {
			return orgID, 0, nil
		}
		if roleCode == proposal.RoleTeacher || roleCode == proposal.RoleCampusAdmin {
			if campusID <= 0 {
				return 0, 0, apperr.InvalidArgument("campusId is required")
			}
		}
		return orgID, campusID, nil
	case actor.IsCampusAdmin():
		if roleCode != proposal.RoleTeacher {
			return 0, 0, apperr.Forbidden("campus admin can only manage teachers")
		}
		if actor.OrganizationID <= 0 || actor.CampusID <= 0 {
			return 0, 0, apperr.Forbidden("campus admin missing campus scope")
		}
		return actor.OrganizationID, actor.CampusID, nil
	default:
		return 0, 0, apperr.Forbidden("admin permission is required")
	}
}

func ensureActorCanManageExistingUser(actor proposal.SessionUserInfo, target *sysUser) error {
	if target == nil {
		return nil
	}
	if actor.IsPlatformAdmin() {
		return nil
	}
	if actor.IsOrgAdmin() {
		if target.RoleCode == proposal.RolePlatformAdmin {
			return apperr.Forbidden("org admin cannot manage platform admins")
		}
		if derefInt32(target.OrganizationID) != actor.OrganizationID {
			return apperr.Forbidden("organization scope mismatch")
		}
		return nil
	}
	if actor.IsCampusAdmin() {
		if target.RoleCode != proposal.RoleTeacher {
			return apperr.Forbidden("campus admin can only manage teachers")
		}
		if derefInt32(target.OrganizationID) != actor.OrganizationID || derefInt32(target.CampusID) != actor.CampusID {
			return apperr.Forbidden("campus scope mismatch")
		}
		return nil
	}
	return apperr.Forbidden("admin permission is required")
}
