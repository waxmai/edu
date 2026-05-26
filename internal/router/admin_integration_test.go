package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	apiAdmin "edu-schedule-system/internal/api/admin"
	"edu-schedule-system/internal/pkg/jwtoken"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/repository/mysql/model"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestAdminGeneratedRoutesIntegration(t *testing.T) {
	if os.Getenv("ADMIN_ROUTE_INTEGRATION_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestAdminGeneratedRoutesIntegration", "-test.v")
		cmd.Env = append(os.Environ(),
			"ADMIN_ROUTE_INTEGRATION_CHILD=1",
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
	adminSvc := newIntegrationAdminService()
	mux, err := NewHTTPMux(
		zap.NewNop(),
		&fakeDBRepo{},
		&fakeCacheRepo{},
		apiAdmin.New(zap.NewNop(), adminSvc),
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

	createBody := requestJSON(t, map[string]string{"username": " demo_admin ", "mobile": "10000000000"})
	createResp := performJSON(t, mux, http.MethodPost, "/api/v1/admin", createBody)
	assertHTTPStatus(t, createResp, http.StatusOK)
	assertJSONDataPathNumber(t, createResp.Body.Bytes(), "id", 1)

	listResp := performJSON(t, mux, http.MethodGet, "/api/v1/admins", nil)
	assertHTTPStatus(t, listResp, http.StatusOK)
	var list dto.AdminListResponse
	decodeJSON(t, responseDataJSON(t, listResp.Body.Bytes()), &list)
	if len(list) != 1 || list[0].Username != "demo_admin" || list[0].Mobile != "10000000000" {
		t.Fatalf("list response = %+v, want created admin", list)
	}

	getResp := performJSON(t, mux, http.MethodGet, "/api/v1/admin/1", nil)
	assertHTTPStatus(t, getResp, http.StatusOK)
	var got dto.AdminResponse
	decodeJSON(t, responseDataJSON(t, getResp.Body.Bytes()), &got)
	if got.ID != 1 || got.Username != "demo_admin" {
		t.Fatalf("get response = %+v, want created admin", got)
	}

	updateBody := requestJSON(t, map[string]string{"mobile": "10000001122"})
	updateResp := performJSON(t, mux, http.MethodPut, "/api/v1/admin/1", updateBody)
	assertHTTPStatus(t, updateResp, http.StatusOK)
	assertJSONDataPathNumber(t, updateResp.Body.Bytes(), "rows_affected", 1)
	updatedResp := performJSON(t, mux, http.MethodGet, "/api/v1/admin/1", nil)
	assertHTTPStatus(t, updatedResp, http.StatusOK)
	var updated dto.AdminResponse
	decodeJSON(t, responseDataJSON(t, updatedResp.Body.Bytes()), &updated)
	if updated.Mobile != "10000001122" {
		t.Fatalf("updated mobile = %q, want %q", updated.Mobile, "10000001122")
	}

	deleteResp := performJSON(t, mux, http.MethodDelete, "/api/v1/admin/1", nil)
	assertHTTPStatus(t, deleteResp, http.StatusOK)
	assertJSONDataPathNumber(t, deleteResp.Body.Bytes(), "rows_affected", 1)

	notFoundResp := performJSON(t, mux, http.MethodGet, "/api/v1/admin/1", nil)
	assertHTTPStatus(t, notFoundResp, http.StatusNotFound)

	badIDResp := performJSON(t, mux, http.MethodGet, "/api/v1/admin/0", nil)
	assertHTTPStatus(t, badIDResp, http.StatusBadRequest)

	badCreateResp := performJSON(t, mux, http.MethodPost, "/api/v1/admin", requestJSON(t, map[string]string{"username": ""}))
	assertHTTPStatus(t, badCreateResp, http.StatusBadRequest)

	conflictResp := performJSON(t, mux, http.MethodPost, "/api/v1/admin", requestJSON(t, map[string]string{"username": "conflict", "mobile": "10000000000"}))
	assertHTTPStatus(t, conflictResp, http.StatusConflict)

	forbiddenResp := performJSON(t, mux, http.MethodPost, "/api/v1/admin", requestJSON(t, map[string]string{"username": "forbidden", "mobile": "10000000000"}))
	assertHTTPStatus(t, forbiddenResp, http.StatusForbidden)

	dependencyResp := performJSON(t, mux, http.MethodPost, "/api/v1/admin", requestJSON(t, map[string]string{"username": "dependency", "mobile": "10000000000"}))
	assertHTTPStatus(t, dependencyResp, http.StatusServiceUnavailable)
}

func performJSON(t *testing.T, handler http.Handler, method, target string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	return performJSONWithSession(t, handler, method, target, body, proposal.SessionUserInfo{})
}

func performJSONWithSession(t *testing.T, handler http.Handler, method, target string, body []byte, session proposal.SessionUserInfo) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		reqBody = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if session.Id > 0 {
		token, err := issueTestToken(session)
		if err != nil {
			t.Fatalf("issueTestToken() error = %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

const routerTestJWTSecret = "router-test-secret-12345678901234567890"

func issueTestToken(session proposal.SessionUserInfo) (string, error) {
	return jwtoken.New(
		routerTestJWTSecret,
		jwtoken.WithIssuer("edu-schedule-system"),
		jwtoken.WithAudience("edu-schedule-system"),
		jwtoken.WithLeeway(0),
	).Sign(session, time.Hour)
}

func requestJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return body
}

func decodeJSON(t *testing.T, body []byte, dst interface{}) {
	t.Helper()
	if err := json.Unmarshal(body, dst); err != nil {
		t.Fatalf("Unmarshal(%s) error = %v", string(body), err)
	}
}

func assertHTTPStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, want, recorder.Body.String())
	}
}

func assertJSONDataPathNumber(t *testing.T, body []byte, key string, want float64) {
	t.Helper()
	var payload map[string]interface{}
	decodeJSON(t, responseDataJSON(t, body), &payload)
	got, ok := payload[key].(float64)
	if !ok {
		t.Fatalf("response %s missing numeric key %q", string(body), key)
	}
	if got != want {
		t.Fatalf("%s = %v, want %v; body=%s", key, got, want, string(body))
	}
}

func responseDataJSON(t *testing.T, body []byte) []byte {
	t.Helper()
	var envelope map[string]interface{}
	decodeJSON(t, body, &envelope)
	data, ok := envelope["data"]
	if !ok {
		t.Fatalf("response %s missing data", string(body))
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal(data) error = %v", err)
	}
	return dataJSON
}

type integrationAdminService struct {
	nextID int32
	items  map[int32]*model.Admin
}

func newIntegrationAdminService() *integrationAdminService {
	return &integrationAdminService{
		nextID: 1,
		items:  map[int32]*model.Admin{},
	}
}

func (s *integrationAdminService) Create(_ context.Context, req *dto.AdminCreateRequest) (int32, error) {
	if req == nil || req.Username == "" || req.Mobile == "" {
		return 0, apperr.InvalidArgument("admin create request is invalid")
	}
	username := trimForIntegration(req.Username)
	switch username {
	case "conflict":
		return 0, apperr.Conflict("admin already exists")
	case "forbidden":
		return 0, apperr.Forbidden("admin create forbidden")
	case "dependency":
		return 0, apperr.DependencyFailed("dependency unavailable")
	}
	id := s.nextID
	s.nextID++
	s.items[id] = &model.Admin{
		ID:        id,
		Username:  username,
		Mobile:    trimForIntegration(req.Mobile),
		CreatedAt: time.Unix(1, 0).UTC(),
	}
	return id, nil
}

func (s *integrationAdminService) List(context.Context) (dto.AdminListResponse, error) {
	list := make(dto.AdminListResponse, 0, len(s.items))
	for id := int32(1); id < s.nextID; id++ {
		if item, ok := s.items[id]; ok {
			list = append(list, adminResponseForIntegration(item))
		}
	}
	return list, nil
}

func (s *integrationAdminService) GetByID(_ context.Context, id int32) (*dto.AdminResponse, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	response := adminResponseForIntegration(item)
	return &response, nil
}

func (s *integrationAdminService) DeleteByID(_ context.Context, id int32) (int64, error) {
	if _, ok := s.items[id]; !ok {
		return 0, gorm.ErrRecordNotFound
	}
	delete(s.items, id)
	return 1, nil
}

func (s *integrationAdminService) UpdateByID(_ context.Context, id int32, req dto.AdminUpdateRequest) (int64, error) {
	item, ok := s.items[id]
	if !ok {
		return 0, gorm.ErrRecordNotFound
	}
	if len(req) == 0 {
		return 0, apperr.InvalidArgument("admin update fields are required")
	}
	for field, value := range req {
		text, ok := value.(string)
		if !ok || text == "" {
			return 0, apperr.InvalidArgument(field + " is invalid")
		}
		switch field {
		case "username":
			item.Username = trimForIntegration(text)
		case "mobile":
			item.Mobile = trimForIntegration(text)
		default:
			return 0, apperr.InvalidArgument("admin update field " + field + " is not allowed")
		}
	}
	return 1, nil
}

func cloneAdmin(item *model.Admin) *model.Admin {
	if item == nil {
		return nil
	}
	copy := *item
	return &copy
}

func adminResponseForIntegration(item *model.Admin) dto.AdminResponse {
	return dto.AdminResponse{
		ID:        item.ID,
		Username:  item.Username,
		Mobile:    item.Mobile,
		CreatedAt: item.CreatedAt,
	}
}

func trimForIntegration(text string) string {
	return strings.TrimSpace(text)
}
