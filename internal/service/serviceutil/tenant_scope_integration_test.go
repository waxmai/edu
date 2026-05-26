package serviceutil

import (
	"testing"

	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
)

func assertForbidden(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	appErr, ok := err.(*apperr.Error)
	if !ok {
		t.Fatalf("error = %T(%v), want *apperr.Error", err, err)
	}
	if appErr.Kind != apperr.KindForbidden {
		t.Fatalf("kind = %q, want %q", appErr.Kind, apperr.KindForbidden)
	}
}

func TestTenantScopeIntegrationRejectsCrossTenantCoreResources(t *testing.T) {
	orgA := proposal.SessionUserInfo{Id: 101, RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeOrg}
	campusA := proposal.SessionUserInfo{Id: 102, RoleCode: proposal.RoleCampusAdmin, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeCampus}
	teacherA := proposal.SessionUserInfo{Id: 103, RoleCode: proposal.RoleTeacher, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeSelf}

	tests := []struct {
		name  string
		actor proposal.SessionUserInfo
		err   error
	}{
		{name: "student cross organization", actor: orgA, err: EnsureStudentAccess(orgA, &model.Student{OrganizationID: 2, CampusID: 20})},
		{name: "course cross organization", actor: orgA, err: EnsureCourseAccess(orgA, &model.Course{OrganizationID: 2, CampusID: 20})},
		{name: "lesson package cross organization", actor: orgA, err: EnsureLessonPackageAccess(orgA, &model.LessonPackage{OrganizationID: 2, CampusID: 20})},
		{name: "payment cross organization", actor: orgA, err: EnsurePaymentRecordAccess(orgA, &model.PaymentRecord{OrganizationID: 2, CampusID: 20})},
		{name: "schedule cross organization", actor: orgA, err: EnsureScheduleAccess(orgA, &model.Schedule{OrganizationID: 2, CampusID: 20, TeacherID: 103})},
		{name: "lesson record cross organization", actor: orgA, err: EnsureLessonRecordAccess(orgA, &model.LessonRecord{OrganizationID: 2, CampusID: 20, TeacherID: 103})},
		{name: "reschedule cross organization", actor: orgA, err: EnsureRescheduleRecordAccess(orgA, &model.RescheduleRecord{OrganizationID: 2, CampusID: 20})},
		{name: "student cross campus", actor: campusA, err: EnsureStudentAccess(campusA, &model.Student{OrganizationID: 1, CampusID: 11})},
		{name: "course cross campus", actor: campusA, err: EnsureCourseAccess(campusA, &model.Course{OrganizationID: 1, CampusID: 11})},
		{name: "teacher schedule owned by another teacher", actor: teacherA, err: EnsureScheduleAccess(teacherA, &model.Schedule{OrganizationID: 1, CampusID: 10, TeacherID: 104})},
		{name: "teacher lesson record owned by another teacher", actor: teacherA, err: EnsureLessonRecordAccess(teacherA, &model.LessonRecord{OrganizationID: 1, CampusID: 10, TeacherID: 104})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertForbidden(t, tt.err)
		})
	}
}

func TestTenantScopeIntegrationAllowsExpectedAccess(t *testing.T) {
	orgA := proposal.SessionUserInfo{Id: 101, RoleCode: proposal.RoleOrgAdmin, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeOrg}
	campusA := proposal.SessionUserInfo{Id: 102, RoleCode: proposal.RoleCampusAdmin, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeCampus}
	teacherA := proposal.SessionUserInfo{Id: 103, RoleCode: proposal.RoleTeacher, OrganizationID: 1, CampusID: 10, DataScope: proposal.DataScopeSelf}
	platform := proposal.SessionUserInfo{Id: 1, RoleCode: proposal.RolePlatformAdmin, DataScope: proposal.DataScopeAll}

	tests := []struct {
		name string
		err  error
	}{
		{name: "org admin same organization different campus", err: EnsureStudentAccess(orgA, &model.Student{OrganizationID: 1, CampusID: 11})},
		{name: "campus admin same campus", err: EnsureCourseAccess(campusA, &model.Course{OrganizationID: 1, CampusID: 10})},
		{name: "teacher own schedule", err: EnsureScheduleAccess(teacherA, &model.Schedule{OrganizationID: 1, CampusID: 10, TeacherID: 103})},
		{name: "teacher own lesson record", err: EnsureLessonRecordAccess(teacherA, &model.LessonRecord{OrganizationID: 1, CampusID: 10, TeacherID: 103})},
		{name: "platform admin bypass", err: EnsurePaymentRecordAccess(platform, &model.PaymentRecord{OrganizationID: 99, CampusID: 88})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				t.Fatalf("unexpected error: %v", tt.err)
			}
		})
	}
}
