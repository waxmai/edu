package interceptor

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/jwtoken"
	"edu-schedule-system/internal/proposal"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type revokedStub struct {
	revoked bool
	err     error
}

func (s revokedStub) Revoke(context.Context, string, time.Duration) error { return nil }
func (s revokedStub) IsRevoked(context.Context, string) (bool, error)     { return s.revoked, s.err }

func TestJWTokenAuthVerifyRejectsRevokedToken(t *testing.T) {
	runJWTVerifyChild(t, "REVOKED")
}

func TestJWTokenAuthVerifyAllowsValidTokenWhenNotRevoked(t *testing.T) {
	runJWTVerifyChild(t, "VALID")
}

func TestJWTokenAuthVerifyRejectsRevokerError(t *testing.T) {
	runJWTVerifyChild(t, "REVOCER_ERROR")
}

func TestJWTokenAuthVerifyChild(t *testing.T) {
	if os.Getenv("JWT_VERIFY_CHILD") != "1" {
		t.Skip("child-only test")
	}

	mode := os.Getenv("JWT_VERIFY_MODE")
	ic := &interceptor{logger: zap.NewNop()}
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())

	switch mode {
	case "REVOKED":
		ic.tokenRevoker = revokedStub{revoked: true}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer revoked-token")
		ginCtx.Request = req
		ctx := core.NewContextForTest(ginCtx)
		if _, err := ic.JWTokenAuthVerify(ctx); err == nil {
			t.Fatal("JWTokenAuthVerify() error = nil, want error")
		}
	case "VALID":
		ic.tokenRevoker = revokedStub{}
		token, signErr := jwtoken.New(
			"test-secret",
			jwtoken.WithIssuer("test-issuer"),
			jwtoken.WithAudience("test-audience"),
		).Sign(proposal.SessionUserInfo{Id: 1, UserName: "admin", Status: proposal.UserStatusEnabled}, time.Hour)
		if signErr != nil {
			t.Fatalf("Sign() error = %v", signErr)
		}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		ginCtx.Request = req
		ctx := core.NewContextForTest(ginCtx)
		session, err := ic.JWTokenAuthVerify(ctx)
		if err != nil {
			t.Fatalf("JWTokenAuthVerify() error = %v", err)
		}
		if session.Id != 1 {
			t.Fatalf("session id = %d, want 1", session.Id)
		}
	case "REVOCER_ERROR":
		ic.tokenRevoker = revokedStub{err: errors.New("redis down")}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer some-token")
		ginCtx.Request = req
		ctx := core.NewContextForTest(ginCtx)
		if _, err := ic.JWTokenAuthVerify(ctx); err == nil {
			t.Fatal("JWTokenAuthVerify() error = nil, want error")
		}
	default:
		t.Fatalf("unknown JWT_VERIFY_MODE %q", mode)
	}
}

func runJWTVerifyChild(t *testing.T, mode string) {
	t.Helper()
	if os.Getenv("JWT_VERIFY_CHILD") == "1" {
		return
	}
	configPath := writeJWTVerifyTestConfig(t)
	cmd := exec.Command(os.Args[0], "-test.run=TestJWTokenAuthVerifyChild", "-test.v")
	cmd.Env = append(os.Environ(),
		"JWT_VERIFY_CHILD=1",
		"JWT_VERIFY_MODE="+mode,
		"CONFIG_PATH="+configPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, string(output))
	}
}

func writeJWTVerifyTestConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	contents := `[mysql.read]
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
secret = "test-secret"
issuer = "test-issuer"
audience = "test-audience"
leewaySeconds = 0

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
