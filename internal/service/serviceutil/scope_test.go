package serviceutil

import (
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
)

func TestEnsureTenantAccessRejectsCrossOrganization(t *testing.T) {
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	if err := EnsureTenantAccess(actor, 2, 1); err == nil {
		t.Fatal("expected cross-organization access to be rejected")
	} else if appErr, ok := err.(*apperr.Error); !ok || appErr.Kind != apperr.KindForbidden {
		t.Fatalf("error = %#v, want forbidden app error", err)
	}
}

func TestEnsureTenantAccessRejectsCrossCampusForCampusScope(t *testing.T) {
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleCampusAdmin, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeCampus}
	if err := EnsureTenantAccess(actor, 1, 20); err == nil {
		t.Fatal("expected cross-campus access to be rejected")
	} else if appErr, ok := err.(*apperr.Error); !ok || appErr.Kind != apperr.KindForbidden {
		t.Fatalf("error = %#v, want forbidden app error", err)
	}
}

func TestEnsureTenantAccessAllowsOrgScopeAcrossCampuses(t *testing.T) {
	actor := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, DataScope: proposal.DataScopeOrg}
	if err := EnsureTenantAccess(actor, 1, 20); err != nil {
		t.Fatalf("EnsureTenantAccess() error = %v, want nil", err)
	}
}

func TestEnsureScheduleAccessRestrictsTeacherToOwnSchedules(t *testing.T) {
	actor := proposal.SessionUserInfo{Id: 7, RoleCode: proposal.RoleTeacher, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeSelf}
	if err := EnsureScheduleAccess(actor, &model.Schedule{OrganizationID: 1, CampusID: 10, TeacherID: 8}); err == nil {
		t.Fatal("expected teacher to be rejected from another teacher schedule")
	} else if appErr, ok := err.(*apperr.Error); !ok || appErr.Kind != apperr.KindForbidden {
		t.Fatalf("error = %#v, want forbidden app error", err)
	}
	if err := EnsureScheduleAccess(actor, &model.Schedule{OrganizationID: 1, CampusID: 10, TeacherID: 7}); err != nil {
		t.Fatalf("EnsureScheduleAccess() error = %v, want nil", err)
	}
}

func TestPlatformAdminBypassesTenantScope(t *testing.T) {
	actor := proposal.SessionUserInfo{RoleCode: proposal.RolePlatformAdmin, DataScope: proposal.DataScopeAll}
	if err := EnsureTenantAccess(actor, 99, 88); err != nil {
		t.Fatalf("EnsureTenantAccess() error = %v, want nil", err)
	}
}
