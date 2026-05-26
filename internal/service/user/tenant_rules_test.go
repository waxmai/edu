package user

import (
	"testing"

	"edu-schedule-system/internal/proposal"
)

func TestEnsureActorCanAssignScope(t *testing.T) {
	platform := proposal.SessionUserInfo{RoleCode: proposal.RolePlatformAdmin, Status: proposal.UserStatusEnabled}
	orgAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10}
	campusAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleCampusAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10, CampusID: 20}

	cases := []struct {
		name      string
		actor     proposal.SessionUserInfo
		roleCode  string
		orgID     int32
		campusID  int32
		wantError bool
	}{
		{name: "platform can assign org admin", actor: platform, roleCode: proposal.RoleOrgAdmin, orgID: 10, campusID: 0, wantError: false},
		{name: "org admin cannot assign platform admin", actor: orgAdmin, roleCode: proposal.RolePlatformAdmin, orgID: 0, campusID: 0, wantError: true},
		{name: "org admin cannot cross organization", actor: orgAdmin, roleCode: proposal.RoleTeacher, orgID: 11, campusID: 20, wantError: true},
		{name: "campus admin can manage teacher in same scope", actor: campusAdmin, roleCode: proposal.RoleTeacher, orgID: 10, campusID: 20, wantError: false},
		{name: "campus admin cannot manage org admin", actor: campusAdmin, roleCode: proposal.RoleOrgAdmin, orgID: 10, campusID: 0, wantError: true},
		{name: "campus admin cannot cross campus", actor: campusAdmin, roleCode: proposal.RoleTeacher, orgID: 10, campusID: 21, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureActorCanAssignScope(tc.actor, tc.roleCode, tc.orgID, tc.campusID)
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestEnsureActorCanManageExistingUser(t *testing.T) {
	platform := proposal.SessionUserInfo{RoleCode: proposal.RolePlatformAdmin, Status: proposal.UserStatusEnabled}
	orgAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10}
	campusAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleCampusAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10, CampusID: 20}
	teacherSameCampus := &sysUser{RoleCode: proposal.RoleTeacher, OrganizationID: nullableInt32(10), CampusID: nullableInt32(20)}
	teacherOtherCampus := &sysUser{RoleCode: proposal.RoleTeacher, OrganizationID: nullableInt32(10), CampusID: nullableInt32(21)}
	platformUser := &sysUser{RoleCode: proposal.RolePlatformAdmin}

	cases := []struct {
		name      string
		actor     proposal.SessionUserInfo
		target    *sysUser
		wantError bool
	}{
		{name: "platform can manage anyone", actor: platform, target: platformUser, wantError: false},
		{name: "org admin cannot manage platform admin", actor: orgAdmin, target: platformUser, wantError: true},
		{name: "org admin can manage same org teacher", actor: orgAdmin, target: teacherSameCampus, wantError: false},
		{name: "campus admin can manage same campus teacher", actor: campusAdmin, target: teacherSameCampus, wantError: false},
		{name: "campus admin cannot manage other campus teacher", actor: campusAdmin, target: teacherOtherCampus, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureActorCanManageExistingUser(tc.actor, tc.target)
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestListRejectsCrossScopeExplicitFilters(t *testing.T) {
	orgAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleOrgAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10, DataScope: proposal.DataScopeOrg}
	campusAdmin := proposal.SessionUserInfo{RoleCode: proposal.RoleCampusAdmin, Status: proposal.UserStatusEnabled, OrganizationID: 10, CampusID: 20, DataScope: proposal.DataScopeCampus}
	service := &service{}

	cases := []struct {
		name  string
		actor proposal.SessionUserInfo
		query ListQuery
	}{
		{name: "org admin cannot filter another organization", actor: orgAdmin, query: ListQuery{OrganizationID: 11}},
		{name: "campus admin cannot filter another campus", actor: campusAdmin, query: ListQuery{OrganizationID: 10, CampusID: 21}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.List(t.Context(), tc.actor, tc.query)
			if err == nil {
				t.Fatal("expected cross-scope filter to be rejected")
			}
		})
	}
}
