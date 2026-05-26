package router

import (
	"net/http"
	"testing"

	"edu-schedule-system/internal/proposal"
)

func TestRouterRoleBoundariesForCourseAndPaymentWrites(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")

	platformAdmin := proposal.SessionUserInfo{Id: 1, UserName: "platform", RoleCode: proposal.RolePlatformAdmin, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RolePlatformAdmin).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RolePlatformAdmin).MenuPermissions, DataScope: proposal.DataScopeAll, FeatureFlags: []string{proposal.FeatureAuth, proposal.FeaturePlatformManagement}, SubscriptionStatus: proposal.SubscriptionStatusActive}
	orgAdmin := proposal.SessionUserInfo{Id: 2, UserName: "org", RoleCode: proposal.RoleOrgAdmin, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleOrgAdmin).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleOrgAdmin).MenuPermissions, DataScope: proposal.DataScopeOrg, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1}
	campusAdmin := proposal.SessionUserInfo{Id: 3, UserName: "campus", RoleCode: proposal.RoleCampusAdmin, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleCampusAdmin).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleCampusAdmin).MenuPermissions, DataScope: proposal.DataScopeCampus, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}

	courseBody := requestJSON(t, map[string]any{
		"courseName":      "权限测试课程",
		"subject":         "数学",
		"courseType":      "group",
		"durationMinutes": 60,
		"feeStandard":     100,
	})
	paymentBody := requestJSON(t, map[string]any{
		"studentId":       1,
		"lessonPackageId": 1,
		"amount":          100,
		"paymentType":     "tuition",
		"paymentMethod":   "cash",
		"paymentTime":     "2026-05-15 10:00:00",
		"paymentStatus":   "paid",
	})

	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/courses", courseBody, teacher), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/payment-records", paymentBody, teacher), http.StatusForbidden)

	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/courses", courseBody, orgAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/courses", courseBody, campusAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/courses", courseBody, platformAdmin), http.StatusForbidden)

	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/payment-records", paymentBody, orgAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/payment-records", paymentBody, campusAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodPost, "/api/v1/payment-records", paymentBody, platformAdmin), http.StatusForbidden)
}
