package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/env"
	redisrepo "edu-schedule-system/internal/repository/redis"

	redisV8 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func authRateLimitMiddleware(logger *zap.Logger, action string, limit int64, window time.Duration, strategy string) core.HandlerFunc {
	if !authEndpointRateLimitEnabled() {
		return func(ctx core.Context) {}
	}

	client, err := redisrepo.GetRedisClient()
	if err != nil {
		if authRateLimitFailOpen() {
			return func(ctx core.Context) {
				if logger != nil {
					logger.Warn("auth rate limit unavailable, fail-open applied", zap.String("action", action), zap.Error(err))
				}
			}
		}
		return func(ctx core.Context) {
			if logger != nil {
				logger.Warn("auth rate limit unavailable", zap.String("action", action), zap.Error(err))
			}
			ctx.AbortWithError(core.Error(503, code.ServerError, code.Text(code.ServerError)))
		}
	}

	prefix := "arl:" + action
	return func(ctx core.Context) {
		key := authRateLimitKey(ctx, strategy)
		if key == "" {
			return
		}
		allowed, retryAfter, allowErr := authRateLimitAllow(ctx.RequestContext(), client, prefix, key, limit, window)
		if allowErr != nil {
			if logger != nil {
				logger.Warn("auth rate limit failed", zap.String("action", action), zap.Error(allowErr))
			}
			ctx.AbortWithError(core.Error(503, code.ServerError, code.Text(code.ServerError)))
			return
		}
		if !allowed {
			if retryAfter > 0 {
				ctx.SetHeader("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
			}
			ctx.AbortWithError(core.Error(429, code.RateLimited, code.Text(code.RateLimited)))
			return
		}
	}
}

func authEndpointRateLimitEnabled() bool {
	value, ok := os.LookupEnv("AUTH_ENDPOINT_RATE_LIMIT_ENABLED")
	if !ok {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "n", "off":
		return false
	default:
		return true
	}
}

func authRateLimitFailOpen() bool {
	if value, ok := os.LookupEnv("AUTH_ENDPOINT_RATE_LIMIT_FAIL_OPEN"); ok {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	if value, ok := os.LookupEnv("REDIS_ENABLED"); ok {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "0", "false", "no", "n", "off":
			return !env.Active().IsPro()
		}
	}
	return false
}

func authRateLimitAllow(ctx context.Context, client *redisV8.Client, prefix, key string, limit int64, window time.Duration) (bool, time.Duration, error) {
	return authRateLimitAllowWithOps(ctx, client.Incr, client.Expire, prefix, key, limit, window)
}

func authRateLimitAllowWithOps(ctx context.Context, incr func(context.Context, string) *redisV8.IntCmd, expire func(context.Context, string, time.Duration) *redisV8.BoolCmd, prefix, key string, limit int64, window time.Duration) (bool, time.Duration, error) {
	if limit <= 0 || window <= 0 {
		return true, 0, nil
	}
	now := time.Now().Unix()
	windowSec := int64(window.Seconds())
	bucket := now / windowSec
	redisKey := fmt.Sprintf("%s:%s:%d", prefix, key, bucket)
	count, err := incr(ctx, redisKey).Result()
	if err != nil {
		return false, 0, err
	}
	ttl := time.Duration((bucket+1)*windowSec-now) * time.Second
	if ttl < 0 {
		ttl = 0
	}
	if count == 1 {
		_ = expire(ctx, redisKey, ttl+time.Second).Err()
	}
	if count > limit {
		return false, ttl, nil
	}
	return true, 0, nil
}

func authRateLimitKey(ctx core.Context, strategy string) string {
	strategy = strings.ToLower(strings.TrimSpace(strategy))
	token := authBearerToken(ctx.GetHeader("Authorization"))
	clientIP := clientIPFromRequest(ctx.Request())
	switch strategy {
	case "token":
		if token != "" {
			return "token:" + authHashToken(token)
		}
		return "ip:" + clientIP
	case "ip":
		return "ip:" + clientIP
	default:
		if token != "" {
			return "token:" + authHashToken(token)
		}
		return "ip:" + clientIP
	}
}

func authBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return ""
}

func authHashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
