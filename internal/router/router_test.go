package router

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"edu-schedule-system/configs"
	apiAdmin "edu-schedule-system/internal/api/admin"
	"edu-schedule-system/internal/service/dto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestRouterAuthMode(t *testing.T) {
	configPath := writeRouterTestConfig(t)

	cases := []struct {
		name       string
		env        string
		authMode   string
		wantStatus string
	}{
		{name: "dev default keeps auth disabled", env: "dev", wantStatus: "200"},
		{name: "pro default keeps auth required", env: "pro", wantStatus: "401"},
		{name: "dev can require auth", env: "dev", authMode: "required", wantStatus: "401"},
	}
	for _, tc := range cases {
		cmd := exec.Command(os.Args[0], "-test.run=TestRouterAuthModeChild", "-test.v")
		cmd.Env = append(os.Environ(),
			"ROUTER_AUTH_CHILD=1",
			"ROUTER_AUTH_ENV="+tc.env,
			"ROUTER_AUTH_WANT_STATUS="+tc.wantStatus,
			"CONFIG_PATH="+configPath,
		)
		if tc.authMode != "" {
			cmd.Env = append(cmd.Env, "AUTH_MODE="+tc.authMode)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s child failed: %v\n%s", tc.name, err, string(output))
		}
	}
}

func TestRouterAuthModeChild(t *testing.T) {
	if os.Getenv("ROUTER_AUTH_CHILD") != "1" {
		t.Skip("child-only test")
	}

	mode := os.Getenv("ROUTER_AUTH_ENV")
	wantStatus := os.Getenv("ROUTER_AUTH_WANT_STATUS")
	os.Args = []string{"router.test", "-env", mode, "-config", os.Getenv("CONFIG_PATH")}
	if _, err := configs.Load(); err != nil {
		t.Fatalf("configs.Load() error = %v", err)
	}

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

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/admins", nil))

	want, err := strconv.Atoi(wantStatus)
	if err != nil {
		t.Fatalf("invalid wanted status %q: %v", wantStatus, err)
	}
	if recorder.Code != want {
		t.Fatalf("mode %s status = %d, want %d; body=%s", mode, recorder.Code, want, recorder.Body.String())
	}
}

func TestRouterTemplateTokenDefaultsDisabledWithoutExplicitOptIn(t *testing.T) {
	if os.Getenv("ROUTER_TEMPLATE_DEFAULT_OFF_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestRouterTemplateTokenDefaultsDisabledWithoutExplicitOptIn", "-test.v")
		cmd.Env = append(os.Environ(),
			"ROUTER_TEMPLATE_DEFAULT_OFF_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	os.Args = []string{"router.test", "-env", "dev", "-config", os.Getenv("CONFIG_PATH")}
	if _, err := configs.Load(); err != nil {
		t.Fatalf("configs.Load() error = %v", err)
	}
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

	assertStatus(t, mux, http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(`{"id":1,"username":"tester"}`), http.StatusNotFound)
}

func TestRouterTemplateTokenDefaultsDisabledInPro(t *testing.T) {
	if os.Getenv("ROUTER_TEMPLATE_PRO_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestRouterTemplateTokenDefaultsDisabledInPro", "-test.v")
		cmd.Env = append(os.Environ(),
			"ROUTER_TEMPLATE_PRO_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	os.Args = []string{"router.test", "-env", "pro", "-config", os.Getenv("CONFIG_PATH")}
	if _, err := configs.Load(); err != nil {
		t.Fatalf("configs.Load() error = %v", err)
	}
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

	assertStatus(t, mux, http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(`{"id":1,"username":"tester"}`), http.StatusNotFound)
}

func TestRouterTemplateInfrastructureChild(t *testing.T) {
	if os.Getenv("ROUTER_TEMPLATE_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestRouterTemplateInfrastructureChild", "-test.v")
		cmd.Env = append(os.Environ(),
			"ROUTER_TEMPLATE_CHILD=1",
			"CONFIG_PATH="+configPath,
			"SWAGGER_ENABLED=true",
			"PPROF_ENABLED=true",
			"METRICS_ENABLED=true",
			"AUTH_TEMPLATE_TOKEN_ENABLED=true",
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	os.Args = []string{"router.test", "-env", "dev", "-config", os.Getenv("CONFIG_PATH")}
	if _, err := configs.Load(); err != nil {
		t.Fatalf("configs.Load() error = %v", err)
	}
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

	assertStatus(t, mux, http.MethodGet, "/metrics", nil, http.StatusOK)
	assertStatus(t, mux, http.MethodGet, "/debug/pprof/", nil, http.StatusOK)
	assertStatus(t, mux, http.MethodGet, "/system/ready", nil, http.StatusOK)
	assertStatus(t, mux, http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(`{"id":1,"username":"tester"}`), http.StatusOK)
	assertStatus(t, mux, http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(`{"id":1,"username":"tester","expire_seconds":7201}`), http.StatusBadRequest)
	assertStatus(t, mux, http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(`{"id":0,"username":"tester"}`), http.StatusBadRequest)
}

func TestRouterReadyReportsDependencyFailure(t *testing.T) {
	if os.Getenv("ROUTER_READY_FAILURE_CHILD") != "1" {
		configPath := writeRouterTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestRouterReadyReportsDependencyFailure", "-test.v")
		cmd.Env = append(os.Environ(),
			"ROUTER_READY_FAILURE_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	os.Args = []string{"router.test", "-env", "dev", "-config", os.Getenv("CONFIG_PATH")}
	if _, err := configs.Load(); err != nil {
		t.Fatalf("configs.Load() error = %v", err)
	}
	mux, err := NewHTTPMux(
		zap.NewNop(),
		&fakeDBRepo{pingErr: errors.New("db down")},
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

	assertStatus(t, mux, http.MethodGet, "/system/ready", nil, http.StatusServiceUnavailable)
}

func assertStatus(t *testing.T, mux http.Handler, method, target string, body *bytes.Buffer, want int) {
	t.Helper()
	assertStatusWithHeaders(t, mux, method, target, bytesOrNil(body), nil, want)
}

func assertStatusWithHeaders(t *testing.T, mux http.Handler, method, target string, body []byte, headers map[string]string, want int) {
	t.Helper()
	recorder := httptest.NewRecorder()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	mux.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, target, recorder.Code, want, recorder.Body.String())
	}
}

func bytesOrNil(buf *bytes.Buffer) []byte {
	if buf == nil {
		return nil
	}
	return buf.Bytes()
}

func writeRouterTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	contents := `
[mysql.read]
addr = "127.0.0.1:3306"
user = "app"
pass = ""
name = "app"

[mysql.write]
addr = "127.0.0.1:3306"
user = "app"
pass = ""
name = "app"

[redis]
enabled = true
addr = "127.0.0.1:6379"
pass = ""
db = 0

[auth]
mode = "auto"

[auth.recoveryDelivery]
mode = "disabled"

[jwt]
secret = "` + routerTestJWTSecret + `"
issuer = "edu-schedule-system"
audience = "edu-schedule-system"
leewaySeconds = 0

[server]
port = ":9999"
maxBodyBytes = 1048576
shutdownTimeoutSeconds = 10
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

type fakeDBRepo struct{ pingErr error }

func (f *fakeDBRepo) GetDbR() *gorm.DB { return nil }
func (f *fakeDBRepo) GetDbW() *gorm.DB { return nil }
func (f *fakeDBRepo) DbRClose() error  { return nil }
func (f *fakeDBRepo) DbWClose() error  { return nil }
func (f *fakeDBRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
func (f *fakeDBRepo) Ping(context.Context) error { return f.pingErr }

type fakeCacheRepo struct{ pingErr error }

func (f *fakeCacheRepo) Set(context.Context, string, interface{}, time.Duration) error { return nil }
func (f *fakeCacheRepo) Get(context.Context, string) (string, error)                   { return "", nil }
func (f *fakeCacheRepo) Del(context.Context, string) error                             { return nil }
func (f *fakeCacheRepo) Close() error                                                  { return nil }
func (f *fakeCacheRepo) Ping(context.Context) error                                    { return f.pingErr }
func (f *fakeCacheRepo) Exists(context.Context, string) (bool, error)                  { return false, nil }
func (f *fakeCacheRepo) GetOrSet(context.Context, string, time.Duration, func() (interface{}, error)) (interface{}, error) {
	return nil, nil
}

type fakeAdminService struct{}

func (fakeAdminService) Create(context.Context, *dto.AdminCreateRequest) (int32, error) {
	return 1, nil
}

func (fakeAdminService) List(context.Context) (dto.AdminListResponse, error) {
	return dto.AdminListResponse{}, nil
}

func (fakeAdminService) GetByID(context.Context, int32) (*dto.AdminResponse, error) {
	return &dto.AdminResponse{ID: 1}, nil
}

func (fakeAdminService) DeleteByID(context.Context, int32) (int64, error) {
	return 1, nil
}

func (fakeAdminService) UpdateByID(context.Context, int32, dto.AdminUpdateRequest) (int64, error) {
	return 1, nil
}
