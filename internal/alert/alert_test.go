package alert

import "testing"

func TestLoadThresholdsFromEnv(t *testing.T) {
	t.Setenv("ALERT_HTTP_LATENCY_SECONDS", "2.5")
	t.Setenv("ALERT_DB_LATENCY_SECONDS", "0.5")
	t.Setenv("ALERT_REDIS_LATENCY_SECONDS", "0.2")
	t.Setenv("ALERT_EXTERNAL_LATENCY_SECONDS", "1.2")
	t.Setenv("ALERT_LOGIN_FAILURES", "9")
	cfg := loadThresholds()
	if cfg.HTTPLatencySeconds != 2.5 || cfg.DBLatencySeconds != 0.5 || cfg.RedisLatencySeconds != 0.2 || cfg.ExternalLatencySeconds != 1.2 || cfg.LoginFailures != 9 {
		t.Fatalf("loadThresholds() = %#v", cfg)
	}
}
