package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"
	"edu-schedule-system/internal/service/apperr"
	"edu-schedule-system/internal/service/dto"

	"go.uber.org/zap"
)

type stubDashboardService struct {
	err   error
	actor proposal.SessionUserInfo
	query dto.DashboardQuery
}

func (s *stubDashboardService) Overview(ctx context.Context, actor proposal.SessionUserInfo, query dto.DashboardQuery) (*dto.DashboardResponse, error) {
	s.actor = actor
	s.query = query
	if s.err != nil {
		return nil, s.err
	}
	return &dto.DashboardResponse{Role: actor.RoleCode}, nil
}

func TestOverviewMapsForbiddenServiceError(t *testing.T) {
	prepareDashboardHandlerTestConfig(t)
	service := &stubDashboardService{err: apperr.Forbidden("actor organization scope is missing")}
	handler := New(service)
	mux, err := core.New(zap.NewNop())
	if err != nil {
		t.Fatalf("core.New() error = %v", err)
	}
	RegisterRoutes(handler, mux.Group("/api/v1"))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/statistics/dashboard", nil)
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

func prepareDashboardHandlerTestConfig(t *testing.T) {
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
secret = "dashboard-test-secret"

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	oldArgs := os.Args
	os.Args = []string{"dashboard.test", "-config", path}
	t.Cleanup(func() { os.Args = oldArgs })
	_, _ = configs.Load()
}
