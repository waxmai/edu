package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"edu-schedule-system/internal/code"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "core-test-config-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	configPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte(coreTestConfigTOML), 0o600); err != nil {
		panic(err)
	}
	_ = os.Setenv("CONFIG_PATH", configPath)

	os.Exit(m.Run())
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(securityHeadersMiddleware())
	engine.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assertHeader(t, recorder, "X-Content-Type-Options", "nosniff")
	assertHeader(t, recorder, "X-Frame-Options", "DENY")
	assertHeader(t, recorder, "Referrer-Policy", "no-referrer")
	assertHeader(t, recorder, "X-XSS-Protection", "0")
}

func TestMaxBodyBytesMiddlewareRejectsDeclaredOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(maxBodyBytesMiddleware(4))
	engine.POST("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345")))

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusRequestEntityTooLarge)
	}
	if !strings.Contains(recorder.Body.String(), code.Text(code.PayloadTooLarge)) {
		t.Fatalf("body = %q, want payload-too-large message", recorder.Body.String())
	}
}

func assertHeader(t *testing.T, recorder *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	if got := recorder.Header().Get(key); got != want {
		t.Fatalf("header %s = %q, want %q", key, got, want)
	}
}

const coreTestConfigTOML = `
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
addr = "127.0.0.1:6379"
pass = ""
db = 0

[jwt]
secret = "core-test-secret"
issuer = "edu-schedule-system"
audience = "edu-schedule-system"
leewaySeconds = 0

[language]
local = "en"

[server]
port = ":9999"
maxBodyBytes = 1048576
shutdownTimeoutSeconds = 10
`
