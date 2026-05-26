package auth

import (
	"context"
	"testing"
	"time"

	redisV8 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type fakeIntCmdable struct {
	counts map[string]int64
	ttl    map[string]time.Duration
}

func newFakeIntCmdable() *fakeIntCmdable {
	return &fakeIntCmdable{counts: map[string]int64{}, ttl: map[string]time.Duration{}}
}

func (f *fakeIntCmdable) Incr(ctx context.Context, key string) *redisV8.IntCmd {
	f.counts[key]++
	return redisV8.NewIntResult(f.counts[key], nil)
}

func (f *fakeIntCmdable) Expire(ctx context.Context, key string, expiration time.Duration) *redisV8.BoolCmd {
	f.ttl[key] = expiration
	return redisV8.NewBoolResult(true, nil)
}

func TestAuthRateLimitAllowBlocksAfterLimit(t *testing.T) {
	client := newFakeIntCmdable()
	prefix := "arl:test"
	key := "ip:198.51.100.8"
	for i := 0; i < 3; i++ {
		allowed, retryAfter, err := authRateLimitAllowWithOps(context.Background(), client.Incr, client.Expire, prefix, key, 3, time.Minute)
		if err != nil {
			t.Fatalf("allow call %d error = %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("allow call %d blocked too early, retryAfter=%v", i+1, retryAfter)
		}
	}
	allowed, retryAfter, err := authRateLimitAllowWithOps(context.Background(), client.Incr, client.Expire, prefix, key, 3, time.Minute)
	if err != nil {
		t.Fatalf("blocked call error = %v", err)
	}
	if allowed {
		t.Fatal("fourth call allowed, want blocked")
	}
	if retryAfter <= 0 {
		t.Fatalf("retryAfter = %v, want > 0", retryAfter)
	}
}

func TestAuthRateLimitAllowSetsExpiryOnFirstHit(t *testing.T) {
	client := newFakeIntCmdable()
	prefix := "arl:test"
	key := "ip:203.0.113.10"
	allowed, _, err := authRateLimitAllowWithOps(context.Background(), client.Incr, client.Expire, prefix, key, 2, time.Minute)
	if err != nil {
		t.Fatalf("first call error = %v", err)
	}
	if !allowed {
		t.Fatal("first call blocked, want allowed")
	}
	if len(client.ttl) != 1 {
		t.Fatalf("ttl writes = %d, want 1", len(client.ttl))
	}
	for _, ttl := range client.ttl {
		if ttl <= 0 {
			t.Fatalf("ttl = %v, want > 0", ttl)
		}
	}
}

func TestAuthRateLimitMiddlewareDisabledByEnv(t *testing.T) {
	t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_ENABLED", "false")
	mw := authRateLimitMiddleware(zap.NewNop(), "login", 10, time.Minute, "ip")
	if mw == nil {
		t.Fatal("authRateLimitMiddleware() = nil, want no-op middleware")
	}
}

func TestAuthRateLimitEnabledByDefault(t *testing.T) {
	t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_ENABLED", "")
	if !authEndpointRateLimitEnabled() {
		t.Fatal("authEndpointRateLimitEnabled() = false, want true by default")
	}
}

func TestAuthRateLimitCanBeDisabledByEnv(t *testing.T) {
	t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_ENABLED", "false")
	if authEndpointRateLimitEnabled() {
		t.Fatal("authEndpointRateLimitEnabled() = true, want false when env disables it")
	}
}

func TestAuthRateLimitFailOpenInDevWithoutRedis(t *testing.T) {
	t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_FAIL_OPEN", "")
	t.Setenv("REDIS_ENABLED", "false")
	if !authRateLimitFailOpen() {
		t.Fatal("authRateLimitFailOpen() = false, want true in dev without redis")
	}
}

func TestAuthRateLimitFailOpenCanBeOverridden(t *testing.T) {
	t.Setenv("AUTH_ENDPOINT_RATE_LIMIT_FAIL_OPEN", "false")
	t.Setenv("REDIS_ENABLED", "false")
	if authRateLimitFailOpen() {
		t.Fatal("authRateLimitFailOpen() = true, want false when env disables it")
	}
}
