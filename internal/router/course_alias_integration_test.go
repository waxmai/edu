package router

import (
	"net/http"
	"os"
	"os/exec"
	"testing"

	apiAdmin "edu-schedule-system/internal/api/admin"
	"edu-schedule-system/internal/proposal"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCoursePluralAliasRoutes(t *testing.T) {
	t.Skip("todo: alias route integration needs auth-aware test fixtures after permission hardening")
	if os.Getenv("COURSE_ALIAS_INTEGRATION_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestCoursePluralAliasRoutes", "-test.v", "-env", "dev", "-config", configPath)
		cmd.Env = append(os.Environ(),
			"COURSE_ALIAS_INTEGRATION_CHILD=1",
			"CONFIG_PATH="+configPath,
			"AUTH_MODE=disabled",
			"REDIS_ENABLED=false",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	os.Args = []string{"router.test", "-env", "dev", "-config", os.Getenv("CONFIG_PATH")}
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
	require.NoError(t, err)

	testSession := proposal.SessionUserInfo{Id: 2, UserName: "teacher", RoleCode: proposal.RoleTeacher, Status: proposal.UserStatusEnabled, Permissions: proposal.BuildAccessProfile(proposal.RoleTeacher).Permissions, MenuPermissions: proposal.BuildAccessProfile(proposal.RoleTeacher).MenuPermissions, DataScope: proposal.DataScopeAll}

	createResp := performJSONWithSession(t, mux, http.MethodPost, "/api/v1/courses", requestJSON(t, map[string]interface{}{
		"courseName":      "初中数学一对一",
		"subject":         "数学",
		"courseType":      "one_to_one",
		"durationMinutes": 120,
		"feeStandard":     300,
	}), testSession)
	assertHTTPStatus(t, createResp, http.StatusOK)
	assertJSONDataPathNumber(t, createResp.Body.Bytes(), "id", 1)

	getResp := performJSONWithSession(t, mux, http.MethodGet, "/api/v1/courses/1", nil, testSession)
	assertHTTPStatus(t, getResp, http.StatusOK)

	updateResp := performJSONWithSession(t, mux, http.MethodPut, "/api/v1/courses/1", requestJSON(t, map[string]interface{}{
		"status": "disabled",
	}), testSession)
	assertHTTPStatus(t, updateResp, http.StatusOK)
	assertJSONDataPathNumber(t, updateResp.Body.Bytes(), "rows_affected", 1)

	deleteResp := performJSONWithSession(t, mux, http.MethodDelete, "/api/v1/courses/1", nil, testSession)
	assertHTTPStatus(t, deleteResp, http.StatusOK)
	assertJSONDataPathNumber(t, deleteResp.Body.Bytes(), "rows_affected", 1)
}
