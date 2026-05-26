package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"edu-schedule-system/configs"
	pkgAudit "edu-schedule-system/internal/pkg/audit"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/service/apperr"

	"go.uber.org/zap"
)

type stubQueryService struct {
	resp       *pkgAudit.ListResponse
	err        error
	query      pkgAudit.Query
	meta       map[string][]string
	exportRows []pkgAudit.ExportRow
}

func (s *stubQueryService) List(query pkgAudit.Query) (*pkgAudit.ListResponse, error) {
	s.query = query
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func (s *stubQueryService) Meta() map[string][]string {
	if s.meta != nil {
		return s.meta
	}
	return map[string][]string{}
}

func (s *stubQueryService) Export(query pkgAudit.Query) ([]pkgAudit.ExportRow, error) {
	s.query = query
	if s.err != nil {
		return nil, s.err
	}
	return s.exportRows, nil
}

func TestHandlerListBindsQueryAndReturnsPayload(t *testing.T) {
	prepareAuditHandlerTestConfig(t)
	service := &stubQueryService{resp: &pkgAudit.ListResponse{Items: []pkgAudit.Record{{Action: "auth.login", Module: "auth"}}, Total: 1, Offset: 0, Limit: 20}}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs?action=auth.login&module=auth&limit=20", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	if service.query.Action != "auth.login" || service.query.Module != "auth" || service.query.Limit != 20 {
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

func TestHandlerListMapsServiceError(t *testing.T) {
	prepareAuditHandlerTestConfig(t)
	service := &stubQueryService{err: apperr.InvalidArgument("bad query")}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs?limit=10", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlerMetaReturnsPayload(t *testing.T) {
	prepareAuditHandlerTestConfig(t)
	service := &stubQueryService{meta: map[string][]string{"actions": {"auth.login"}, "modules": {"auth"}}}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs/meta", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload["code"].(float64) != 0 {
		t.Fatalf("response code = %#v, want 0", payload["code"])
	}
}

func TestHandlerExportReturnsCSVAttachment(t *testing.T) {
	prepareAuditHandlerTestConfig(t)
	service := &stubQueryService{exportRows: []pkgAudit.ExportRow{{Time: "2026-05-15T10:00:00+08:00", Action: "auth.login", Module: "auth", ActorUsername: "admin", Message: "audit-event"}}}
	handler := New(zap.NewNop(), service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	group := mux.Group("/api/v1")
	RegisterRoutes(handler, group, group)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs/export?action=auth.login&limit=100", nil)
	mux.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	if service.query.Action != "auth.login" || service.query.Limit != 100 {
		t.Fatalf("bound export query = %#v", service.query)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType == "" || contentType[:8] != "text/csv" {
		t.Fatalf("content type = %q, want text/csv", contentType)
	}
	if disposition := recorder.Header().Get("Content-Disposition"); disposition == "" {
		t.Fatalf("content disposition empty")
	}
	body := recorder.Body.String()
	if body == "" || body[:4] != "time" {
		t.Fatalf("csv body = %q", body)
	}
}

func prepareAuditHandlerTestConfig(t *testing.T) {
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
secret = "audit-test-secret"

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"audit.test", "-config", path}
	t.Cleanup(func() { os.Args = oldArgs })
	_, _ = configs.Load()
}
