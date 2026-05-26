package configs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"edu-schedule-system/internal/pkg/env"
)

func TestLoadPrefersConfigFlagOverEnv(t *testing.T) {
	flagConfig := writeTestConfig(t, "flag-secret", "flag-db")
	envConfig := writeTestConfig(t, "env-secret", "env-db")

	withConfigTestState(t, []string{"test", "-config", flagConfig}, map[string]string{"CONFIG_PATH": envConfig})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.Secret != "flag-secret" {
		t.Fatalf("JWT.Secret = %q, want flag-secret", cfg.JWT.Secret)
	}
	if cfg.MySQL.Read.Name != "flag-db" {
		t.Fatalf("MySQL.Read.Name = %q, want flag-db", cfg.MySQL.Read.Name)
	}
}

func TestLoadUsesConfigPathEnv(t *testing.T) {
	envConfig := writeTestConfig(t, "env-secret", "env-db")

	withConfigTestState(t, []string{"test"}, map[string]string{"CONFIG_PATH": envConfig})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.Secret != "env-secret" {
		t.Fatalf("JWT.Secret = %q, want env-secret", cfg.JWT.Secret)
	}
}

func TestLoadFindsDefaultConfigFromParentDirectory(t *testing.T) {
	root := t.TempDir()
	configsDir := filepath.Join(root, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	defaultPath := filepath.Join(configsDir, "fat_configs.toml")
	if err := os.WriteFile(defaultPath, []byte(testConfigTOML("default-secret", "default-db")), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	withConfigTestState(t, []string{"test"}, map[string]string{"CONFIG_PATH": ""})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.Secret != "default-secret" {
		t.Fatalf("JWT.Secret = %q, want default-secret", cfg.JWT.Secret)
	}
}

func TestLoadRejectsMissingJWTSecret(t *testing.T) {
	configPath := writeTestConfig(t, "", "app")

	withConfigTestState(t, []string{"test", "-config", configPath}, nil)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want jwt secret validation error")
	}
	if !strings.Contains(err.Error(), "jwt.secret is required") {
		t.Fatalf("Load() error = %q, want jwt.secret validation", err.Error())
	}
}

func TestLoadAllowsNestedEnvOverrides(t *testing.T) {
	configPath := writeTestConfig(t, "file-secret", "file-db")

	withConfigTestState(t, []string{"test", "-config", configPath}, map[string]string{
		"JWT_SECRET":                  "env-secret",
		"MYSQL_READ_ADDR":             "read-db:3306",
		"MYSQL_WRITE_NAME":            "write-db",
		"REDIS_ADDR":                  "redis:6379",
		"REDIS_ENABLED":               "false",
		"AUTH_MODE":                   "required",
		"AUTH_TEMPLATE_TOKEN_ENABLED": "false",
		"AUTH_RECOVERY_DELIVERY_MODE": "disabled",
		"SERVER_PORT":                 "8080",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.Secret != "env-secret" {
		t.Fatalf("JWT.Secret = %q, want env-secret", cfg.JWT.Secret)
	}
	if cfg.MySQL.Read.Addr != "read-db:3306" {
		t.Fatalf("MySQL.Read.Addr = %q, want read-db:3306", cfg.MySQL.Read.Addr)
	}
	if cfg.MySQL.Write.Name != "write-db" {
		t.Fatalf("MySQL.Write.Name = %q, want write-db", cfg.MySQL.Write.Name)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Fatalf("Redis.Addr = %q, want redis:6379", cfg.Redis.Addr)
	}
	if cfg.Redis.Enabled {
		t.Fatal("Redis.Enabled = true, want false")
	}
	if cfg.Auth.Mode != "required" {
		t.Fatalf("Auth.Mode = %q, want required", cfg.Auth.Mode)
	}
	if !cfg.Auth.TemplateTokenConfigured || cfg.Auth.TemplateTokenEnabled {
		t.Fatal("Auth.TemplateTokenEnabled = true/unconfigured, want configured false")
	}
	if cfg.Server.Port != ":8080" {
		t.Fatalf("Server.Port = %q, want :8080", cfg.Server.Port)
	}
	if cfg.Auth.RecoveryDelivery.Mode != "disabled" {
		t.Fatalf("Auth.RecoveryDelivery.Mode = %q, want disabled", cfg.Auth.RecoveryDelivery.Mode)
	}
}

func TestLoadAllowsMissingRedisAddrWhenDisabled(t *testing.T) {
	configPath := writeRawConfig(t, `
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
secret = "secret"

[server]
port = ":9999"
`)

	withConfigTestState(t, []string{"test", "-config", configPath}, nil)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Redis.Enabled {
		t.Fatal("Redis.Enabled = true, want false")
	}
}

func TestLoadRejectsDisabledAuthModeInPro(t *testing.T) {
	if os.Getenv("CONFIG_PRO_AUTH_DISABLED_CHILD") != "1" {
		configPath := writeRawConfig(t, strings.Replace(testConfigTOML("prod-secret-for-test-12345678901234567890", "app"), "[jwt]", "[auth]\nmode = \"disabled\"\n\n[jwt]", 1))
		cmd := exec.Command(os.Args[0], "-test.run=TestLoadRejectsDisabledAuthModeInPro", "-test.v")
		cmd.Env = append(os.Environ(),
			"CONFIG_PRO_AUTH_DISABLED_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	withConfigTestState(t, []string{"test", "-env", "pro"}, map[string]string{"CONFIG_PATH": os.Getenv("CONFIG_PATH")})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want pro auth disabled validation error")
	}
	if !strings.Contains(err.Error(), "auth.mode must not be disabled in pro") {
		t.Fatalf("Load() error = %q, want auth.mode pro validation", err.Error())
	}
}

func TestLoadRejectsWeakJWTSecretInPro(t *testing.T) {
	if os.Getenv("CONFIG_PRO_WEAK_JWT_CHILD") != "1" {
		configPath := writeTestConfig(t, "dev-secret-local-only-change-me", "app")
		cmd := exec.Command(os.Args[0], "-test.run=TestLoadRejectsWeakJWTSecretInPro", "-test.v")
		cmd.Env = append(os.Environ(),
			"CONFIG_PRO_WEAK_JWT_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	withConfigTestState(t, []string{"test", "-env", "pro"}, map[string]string{"CONFIG_PATH": os.Getenv("CONFIG_PATH")})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want weak jwt secret validation error")
	}
	if !strings.Contains(err.Error(), "jwt.secret must be a strong non-default value in pro") {
		t.Fatalf("Load() error = %q, want weak jwt pro validation", err.Error())
	}
}

func TestLoadRejectsTemplateTokenInPro(t *testing.T) {
	if os.Getenv("CONFIG_PRO_TEMPLATE_CHILD") != "1" {
		configPath := writeRawConfig(t, strings.Replace(testConfigTOML("prod-secret-for-test-12345678901234567890", "app"), "[jwt]", "[auth]\ntemplateTokenEnabled = true\n\n[jwt]", 1))
		cmd := exec.Command(os.Args[0], "-test.run=TestLoadRejectsTemplateTokenInPro", "-test.v")
		cmd.Env = append(os.Environ(),
			"CONFIG_PRO_TEMPLATE_CHILD=1",
			"CONFIG_PATH="+configPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	withConfigTestState(t, []string{"test", "-env", "pro"}, map[string]string{"CONFIG_PATH": os.Getenv("CONFIG_PATH")})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want pro template token validation error")
	}
	if !strings.Contains(err.Error(), "auth.templateTokenEnabled must be false in pro") {
		t.Fatalf("Load() error = %q, want template token pro validation", err.Error())
	}
}

func TestLoadRejectsPlaceholderRecoveryDeliveryInPro(t *testing.T) {
	if os.Getenv("CONFIG_PRO_RECOVERY_PLACEHOLDER_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=TestLoadRejectsPlaceholderRecoveryDeliveryInPro", "-test.v", "-env", "pro")
		cmd.Env = append(os.Environ(), "CONFIG_PRO_RECOVERY_PLACEHOLDER_CHILD=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("child failed: %v\n%s", err, string(output))
		}
		return
	}

	env.Init()
	cfg := Config{}
	cfg.JWT.Secret = strings.Repeat("test", 10)
	cfg.MySQL.Read.Addr = "127.0.0.1:3306"
	cfg.MySQL.Read.User = "app"
	cfg.MySQL.Read.Name = "app"
	cfg.MySQL.Write.Addr = "127.0.0.1:3306"
	cfg.MySQL.Write.User = "app"
	cfg.MySQL.Write.Name = "app"
	cfg.Redis.Enabled = false
	cfg.Server.Port = ":9999"
	cfg.Auth.Mode = "required"
	cfg.Auth.RecoveryDelivery.Mode = "placeholder"

	if err := validateConfig(&cfg); err == nil {
		t.Fatal("validateConfig() error = nil, want pro recovery delivery validation error")
	} else if !strings.Contains(err.Error(), "auth.recoveryDelivery.mode must not be placeholder in pro") {
		t.Fatalf("validateConfig() error = %q, want recovery delivery pro validation", err.Error())
	}
}

func TestLoadRejectsEmailRecoveryDeliveryWithoutRequiredConfig(t *testing.T) {
	cfg := Config{}
	cfg.JWT.Secret = "secret"
	cfg.MySQL.Read.Addr = "127.0.0.1:3306"
	cfg.MySQL.Read.User = "app"
	cfg.MySQL.Read.Name = "app"
	cfg.MySQL.Write.Addr = "127.0.0.1:3306"
	cfg.MySQL.Write.User = "app"
	cfg.MySQL.Write.Name = "app"
	cfg.Redis.Enabled = false
	cfg.Server.Port = ":9999"
	cfg.Auth.Mode = "required"
	cfg.Auth.RecoveryDelivery.Mode = "email"

	withConfigTestState(t, []string{"test", "-env", "fat"}, nil)
	env.Init()
	if err := validateConfig(&cfg); err == nil {
		t.Fatal("validateConfig() error = nil, want missing email delivery config failure")
	} else if !strings.Contains(err.Error(), "auth.recoveryDelivery.email.host is required") {
		t.Fatalf("validateConfig() error = %q, want email host validation", err.Error())
	}
}

func TestLoadAcceptsEmailRecoveryDeliveryWhenConfigured(t *testing.T) {
	cfg := Config{}
	cfg.JWT.Secret = "secret"
	cfg.MySQL.Read.Addr = "127.0.0.1:3306"
	cfg.MySQL.Read.User = "app"
	cfg.MySQL.Read.Name = "app"
	cfg.MySQL.Write.Addr = "127.0.0.1:3306"
	cfg.MySQL.Write.User = "app"
	cfg.MySQL.Write.Name = "app"
	cfg.Redis.Enabled = false
	cfg.Server.Port = ":9999"
	cfg.Auth.Mode = "required"
	cfg.Auth.RecoveryDelivery.Mode = "email"
	cfg.Auth.RecoveryDelivery.Email.Host = "smtp.example.com"
	cfg.Auth.RecoveryDelivery.Email.Port = 587
	cfg.Auth.RecoveryDelivery.Email.Username = "mailer"
	cfg.Auth.RecoveryDelivery.Email.Password = "secret-pass"
	cfg.Auth.RecoveryDelivery.Email.From = "noreply@example.com"

	withConfigTestState(t, []string{"test", "-env", "fat"}, nil)
	env.Init()
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig() error = %v, want nil", err)
	}
}

func TestLoadAcceptsSMSRecoveryDeliveryWhenConfigured(t *testing.T) {
	cfg := Config{}
	cfg.JWT.Secret = "secret"
	cfg.MySQL.Read.Addr = "127.0.0.1:3306"
	cfg.MySQL.Read.User = "app"
	cfg.MySQL.Read.Name = "app"
	cfg.MySQL.Write.Addr = "127.0.0.1:3306"
	cfg.MySQL.Write.User = "app"
	cfg.MySQL.Write.Name = "app"
	cfg.Redis.Enabled = false
	cfg.Server.Port = ":9999"
	cfg.Auth.Mode = "required"
	cfg.Auth.RecoveryDelivery.Mode = "sms"
	cfg.Auth.RecoveryDelivery.SMS.Provider = "stub"

	withConfigTestState(t, []string{"test", "-env", "fat"}, nil)
	env.Init()
	if err := validateConfig(&cfg); err != nil {
		t.Fatalf("validateConfig() error = %v, want nil", err)
	}
}

func TestLoadRejectsSMSRecoveryDeliveryWithoutProvider(t *testing.T) {
	cfg := Config{}
	cfg.JWT.Secret = "secret"
	cfg.MySQL.Read.Addr = "127.0.0.1:3306"
	cfg.MySQL.Read.User = "app"
	cfg.MySQL.Read.Name = "app"
	cfg.MySQL.Write.Addr = "127.0.0.1:3306"
	cfg.MySQL.Write.User = "app"
	cfg.MySQL.Write.Name = "app"
	cfg.Redis.Enabled = false
	cfg.Server.Port = ":9999"
	cfg.Auth.Mode = "required"
	cfg.Auth.RecoveryDelivery.Mode = "sms"

	withConfigTestState(t, []string{"test", "-env", "fat"}, nil)
	env.Init()
	if err := validateConfig(&cfg); err == nil {
		t.Fatal("validateConfig() error = nil, want missing sms provider failure")
	} else if !strings.Contains(err.Error(), "auth.recoveryDelivery.sms.provider is required") {
		t.Fatalf("validateConfig() error = %q, want sms provider validation", err.Error())
	}
}

func TestLoadDefaultsShutdownTimeout(t *testing.T) {
	configPath := writeTestConfig(t, "secret", "app")

	withConfigTestState(t, []string{"test", "-config", configPath}, nil)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.ShutdownTimeoutSeconds != 10 {
		t.Fatalf("ShutdownTimeoutSeconds = %d, want 10", cfg.Server.ShutdownTimeoutSeconds)
	}
	if cfg.Server.ReadHeaderTimeoutSeconds != 5 {
		t.Fatalf("ReadHeaderTimeoutSeconds = %d, want 5", cfg.Server.ReadHeaderTimeoutSeconds)
	}
	if cfg.Server.ReadTimeoutSeconds != 10 {
		t.Fatalf("ReadTimeoutSeconds = %d, want 10", cfg.Server.ReadTimeoutSeconds)
	}
	if cfg.Server.WriteTimeoutSeconds != 15 {
		t.Fatalf("WriteTimeoutSeconds = %d, want 15", cfg.Server.WriteTimeoutSeconds)
	}
	if cfg.Server.IdleTimeoutSeconds != 60 {
		t.Fatalf("IdleTimeoutSeconds = %d, want 60", cfg.Server.IdleTimeoutSeconds)
	}
	if cfg.MySQL.Pool.MaxOpenConns != 100 {
		t.Fatalf("MySQL.Pool.MaxOpenConns = %d, want 100", cfg.MySQL.Pool.MaxOpenConns)
	}
	if cfg.MySQL.Pool.MaxIdleConns != 5 {
		t.Fatalf("MySQL.Pool.MaxIdleConns = %d, want 5", cfg.MySQL.Pool.MaxIdleConns)
	}
	if cfg.MySQL.Pool.ConnMaxLifetimeSeconds != 120 {
		t.Fatalf("MySQL.Pool.ConnMaxLifetimeSeconds = %d, want 120", cfg.MySQL.Pool.ConnMaxLifetimeSeconds)
	}
	if cfg.Redis.Pool.PoolSize != 50 {
		t.Fatalf("Redis.Pool.PoolSize = %d, want 50", cfg.Redis.Pool.PoolSize)
	}
}

func TestLoadRejectsMissingDependencyConfig(t *testing.T) {
	configPath := writeRawConfig(t, `
[mysql.read]
addr = ""
user = "app"
name = "app"

[mysql.write]
addr = "127.0.0.1:3306"
user = "app"
name = "app"

[redis]
addr = "127.0.0.1:6379"

[jwt]
secret = "secret"

[server]
port = ":9999"
`)

	withConfigTestState(t, []string{"test", "-config", configPath}, nil)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want mysql.read.addr validation error")
	}
	if !strings.Contains(err.Error(), "mysql.read.addr is required") {
		t.Fatalf("Load() error = %q, want mysql.read.addr validation", err.Error())
	}
}

func withConfigTestState(t *testing.T, args []string, env map[string]string) {
	t.Helper()

	oldArgs := os.Args
	oldConfig := config
	oldErr := configErr
	trackedEnv := []string{"CONFIG_PATH", "JWT_SECRET", "MYSQL_READ_ADDR", "MYSQL_WRITE_NAME", "REDIS_ADDR", "REDIS_ENABLED", "AUTH_MODE", "AUTH_TEMPLATE_TOKEN_ENABLED", "AUTH_RECOVERY_DELIVERY_MODE", "SERVER_PORT"}
	oldEnv := make(map[string]string, len(trackedEnv))
	hadEnv := make(map[string]bool, len(trackedEnv))
	for _, key := range trackedEnv {
		oldEnv[key], hadEnv[key] = os.LookupEnv(key)
	}

	os.Args = args
	config = Config{}
	configOnce = sync.Once{}
	configErr = nil
	if _, ok := env["CONFIG_PATH"]; !ok {
		_ = os.Unsetenv("CONFIG_PATH")
	}
	_ = os.Unsetenv("JWT_SECRET")
	for _, key := range trackedEnv {
		if key != "CONFIG_PATH" {
			_ = os.Unsetenv(key)
		}
	}
	for key, value := range env {
		if value == "" {
			_ = os.Unsetenv(key)
			continue
		}
		_ = os.Setenv(key, value)
	}

	t.Cleanup(func() {
		os.Args = oldArgs
		config = oldConfig
		configOnce = sync.Once{}
		configErr = oldErr
		for _, key := range trackedEnv {
			if hadEnv[key] {
				_ = os.Setenv(key, oldEnv[key])
			} else {
				_ = os.Unsetenv(key)
			}
		}
	})
}

func writeTestConfig(t *testing.T, jwtSecret, dbName string) string {
	t.Helper()

	return writeRawConfig(t, testConfigTOML(jwtSecret, dbName))
}

func writeRawConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func testConfigTOML(jwtSecret, dbName string) string {
	return `
[mysql.read]
addr = "127.0.0.1:3306"
user = "app"
pass = ""
name = "` + dbName + `"

[mysql.write]
addr = "127.0.0.1:3306"
user = "app"
pass = ""
name = "` + dbName + `"

[redis]
addr = "127.0.0.1:6379"
pass = ""
db = 0

[jwt]
secret = "` + jwtSecret + `"
issuer = "edu-schedule-system"
audience = "edu-schedule-system"
leewaySeconds = 0

[server]
port = ":9999"
maxBodyBytes = 1048576
`
}
