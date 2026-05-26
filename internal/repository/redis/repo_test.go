package redis

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCacheReturnsNoopWhenRedisDisabled(t *testing.T) {
	if os.Getenv("REDIS_NOOP_CHILD") != "1" {
		configPath := writeRedisTestConfig(t)
		cmd := exec.Command(os.Args[0], "-test.run=TestNewCacheReturnsNoopWhenRedisDisabled", "-test.v")
		cmd.Env = append(os.Environ(),
			"REDIS_NOOP_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	repo, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	if err := repo.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	value, err := repo.GetOrSet(context.Background(), "key", time.Minute, func() (interface{}, error) {
		return "fresh", nil
	})
	if err != nil {
		t.Fatalf("GetOrSet() error = %v", err)
	}
	if value != "fresh" {
		t.Fatalf("GetOrSet() = %v, want fresh", value)
	}
}

func writeRedisTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	contents := `
[mysql.read]
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
secret = "redis-test-secret"

[server]
port = ":9999"
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
