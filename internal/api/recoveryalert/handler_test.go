package recoveryalert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	pkgRecoveryAlert "edu-schedule-system/internal/pkg/recoveryalert"
	"edu-schedule-system/internal/service/apperr"

	"go.uber.org/zap"
)

type stubQueryService struct {
	resp    *pkgRecoveryAlert.ListResponse
	summary *pkgRecoveryAlert.SummaryResponse
	err     error
	query   pkgRecoveryAlert.Query
}

func (s *stubQueryService) List(query pkgRecoveryAlert.Query) (*pkgRecoveryAlert.ListResponse, error) {
	s.query = query
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func (s *stubQueryService) Summary(query pkgRecoveryAlert.Query) (*pkgRecoveryAlert.SummaryResponse, error) {
	s.query = query
	if s.err != nil {
		return nil, s.err
	}
	if s.summary != nil {
		return s.summary, nil
	}
	return &pkgRecoveryAlert.SummaryResponse{}, nil
}

func TestHandlerListBindsQueryAndReturnsPayload(t *testing.T) {
	prepareRecoveryAlertHandlerTestConfig(t)
	service := &stubQueryService{resp: &pkgRecoveryAlert.ListResponse{Items: []pkgRecoveryAlert.Record{{Username: "alice", Channel: "email"}}, Total: 1, Offset: 0, Limit: 20}}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery-alerts?username=alice&channel=email&limit=20", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	if service.query.Username != "alice" || service.query.Channel != "email" || service.query.Limit != 20 {
		t.Fatalf("bound query = %#v", service.query)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload["code"].(float64) != 0 {
		t.Fatalf("response code = %#v, want 0", payload["code"])
	}
}

func TestHandlerSummaryReturnsPayload(t *testing.T) {
	prepareRecoveryAlertHandlerTestConfig(t)
	service := &stubQueryService{summary: &pkgRecoveryAlert.SummaryResponse{Total: 3, ByChannel: map[string]int{"email": 2}, ByErrorCategory: map[string]int{"timeout": 1}}}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery-alerts/summary?channel=email", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	if service.query.Channel != "email" {
		t.Fatalf("bound summary query = %#v", service.query)
	}
}

func TestHandlerListMapsForbiddenServiceError(t *testing.T) {
	prepareRecoveryAlertHandlerTestConfig(t)
	service := &stubQueryService{err: apperr.Forbidden("admin permission is required")}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	RegisterRoutes(handler, mux.Group("/api/v1"))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recovery-alerts", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload code.Failure
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Code != code.Forbidden {
		t.Fatalf("business code = %d, want %d", payload.Code, code.Forbidden)
	}
}

func prepareRecoveryAlertHandlerTestConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[mysql.read]
addr = "127.0.0.1:3306"
user = "app"
name = "app"

[mysql.write]
addr = "127.0.0.1:3306"
user = "app"
name = "app"

[redis]
enabled = false
addr = ""

[jwt]
secret = "recovery-alert-test-secret"

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"recoveryalert.test", "-config", path}
	t.Cleanup(func() { os.Args = oldArgs })
	_, _ = configs.Load()
}
