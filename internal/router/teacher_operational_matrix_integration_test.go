package router

import (
	"net/http"
	"testing"

	"edu-schedule-system/internal/proposal"
)

func TestRouterTeacherCannotWriteLessonPackage(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	lessonPackageBody := requestJSON(t, map[string]any{
		"studentId":     1,
		"courseId":      1,
		"totalLessons":  20,
		"usedLessons":   0,
		"remainLessons": 20,
		"status":        "active",
	})
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/lesson-packages", lessonPackageBody, teacher), http.StatusForbidden)
}

func TestRouterTeacherCanWriteSchedule(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	scheduleBody := requestJSON(t, map[string]any{
		"studentId":      1,
		"courseId":       1,
		"teacherId":      4,
		"classDate":      "2026-05-16",
		"startTime":      "2026-05-16 10:00:00",
		"endTime":        "2026-05-16 11:00:00",
		"scheduleStatus": "scheduled",
	})
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/schedules", scheduleBody, teacher), http.StatusOK)
}

func TestRouterTeacherCanWriteLessonRecord(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	lessonRecordBody := requestJSON(t, map[string]any{
		"scheduleId":        1,
		"studentId":         1,
		"teacherId":         4,
		"attendanceStatus":  "present",
		"needDeductLesson":  true,
		"deductLessonCount": 1,
	})
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/lesson-records", lessonRecordBody, teacher), http.StatusOK)
}

func TestRouterTeacherCannotWriteRescheduleRecord(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	rescheduleBody := requestJSON(t, map[string]any{
		"oldScheduleId": 1,
		"newScheduleId": 2,
		"operationType": "reschedule",
	})
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/reschedule-records", rescheduleBody, teacher), http.StatusForbidden)
}
