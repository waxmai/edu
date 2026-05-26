package router

import (
	"net/http"
	"os"
	"testing"

	apiAdmin "edu-schedule-system/internal/api/admin"
	"edu-schedule-system/internal/proposal"

	"go.uber.org/zap"
)

func TestRouterRoleBoundariesForPlatformAudit(t *testing.T) {
	mux := newAuditRoleTestMux(t, "required")

	platformAdmin := sessionForRole(1, "platform", proposal.RolePlatformAdmin)
	platformAuditor := sessionForRole(5, "auditor", proposal.RolePlatformAuditor)
	platformFinance := sessionForRole(6, "finance", proposal.RolePlatformFinance)
	orgAdmin := proposal.SessionUserInfo{Id: 2, UserName: "org", RoleCode: proposal.RoleOrgAdmin, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleOrgAdmin).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleOrgAdmin).MenuPermissions, DataScope: proposal.DataScopeOrg, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1}
	campusAdmin := proposal.SessionUserInfo{Id: 3, UserName: "campus", RoleCode: proposal.RoleCampusAdmin, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleCampusAdmin).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleCampusAdmin).MenuPermissions, DataScope: proposal.DataScopeCampus, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}
	teacher := proposal.SessionUserInfo{Id: 4, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeSelf, FeatureFlags: []string{proposal.FeatureAuth}, SubscriptionStatus: proposal.SubscriptionStatusActive, OrganizationID: 1, CampusID: 1}

	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs", nil, platformAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/recovery-alerts", nil, platformAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/notification-records", nil, platformAdmin), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs", nil, platformAuditor), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs/export", nil, platformAuditor), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/notification-records", nil, platformAuditor), http.StatusOK)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs/export", nil, platformFinance), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs", nil, orgAdmin), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/recovery-alerts", nil, orgAdmin), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/notification-records", nil, orgAdmin), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs", nil, campusAdmin), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/recovery-alerts", nil, campusAdmin), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/audit/logs", nil, teacher), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/recovery-alerts", nil, teacher), http.StatusForbidden)
	assertHTTPStatus(t, performJSONWithSession(t, mux, http.MethodGet, "/api/v1/notification-records", nil, teacher), http.StatusForbidden)
}

func sessionForRole(id int32, username string, role string) proposal.SessionUserInfo {
	profile := proposal.BuildAccessProfile(role)
	return proposal.SessionUserInfo{Id: id, UserName: username, RoleCode: role, Status: proposal.UserStatusEnabled, Permissions: profile.Permissions, MenuPermissions: profile.MenuPermissions, DataScope: profile.DataScope, FeatureFlags: []string{proposal.FeatureAuth, proposal.FeaturePlatformManagement, proposal.FeatureSubscriptionCenter, proposal.FeatureAuditExport, proposal.FeatureRecoveryOps, proposal.FeatureCustomerSuccess, proposal.FeaturePlatformReport}, SubscriptionStatus: proposal.SubscriptionStatusActive}
}

func newAuditRoleTestMux(t *testing.T, mode string) http.Handler {
	t.Helper()
	configPath := writeRouterTestConfig(t)
	originalArgs := os.Args
	os.Args = []string{"router.test", "-env", "dev", "-config", configPath}
	t.Cleanup(func() { os.Args = originalArgs })
	t.Setenv("CONFIG_PATH", configPath)
	t.Setenv("AUTH_MODE", mode)
	mux, err := NewHTTPMux(
		zap.NewNop(),
		&fakeDBRepo{},
		&fakeCacheRepo{},
		apiAdmin.New(zap.NewNop(), fakeAdminService{}),
		newFakeStudentHandler(),
		newFakeCourseHandler(),
		newFakeLessonPackageHandler(),
		newFakePaymentRecordHandler(),
		newFakeScheduleHandler(),
		newFakeLessonRecordHandler(),
		newFakeRescheduleRecordHandler(),
	)
	if err != nil {
		t.Fatalf("NewHTTPMux() error = %v", err)
	}
	return mux
}
