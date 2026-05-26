package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"edu-schedule-system/internal/code"
	redisrepo "edu-schedule-system/internal/repository/redis"

	"github.com/gin-gonic/gin"
	redisV8 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Config struct {
	Enabled      bool
	Limit        int64
	Window       time.Duration
	Prefix       string
	FailOpen     bool
	KeyStrategy  string
	ExcludePaths []string
}

type redisLimiter struct {
	client *redisV8.Client
	cfg    Config
	logger *zap.Logger
}

func New(logger *zap.Logger, cfg Config) gin.HandlerFunc {
	limiter, blockHandler := buildLimiter(logger, cfg)
	if limiter == nil {
		return blockHandler
	}
	return limiter.middleware()
}

func NewFromEnv(logger *zap.Logger) gin.HandlerFunc {
	return New(logger, configFromEnv())
}

func buildLimiter(logger *zap.Logger, cfg Config) (*redisLimiter, gin.HandlerFunc) {
	if !cfg.Enabled {
		return nil, nil
	}
	if len(cfg.ExcludePaths) == 0 {
		cfg.ExcludePaths = []string{"/system/health"}
	}
	client, err := redisrepo.GetRedisClient()
	if err != nil {
		if cfg.FailOpen {
			if logger != nil {
				logger.Warn("rate limit disabled because redis is unavailable", zap.Error(err))
			}
			return nil, nil
		}
		return nil, func(c *gin.Context) {
			c.AbortWithStatusJSON(httpStatusServiceUnavailable, code.Failure{
				Code:    code.ServerError,
				Message: code.Text(code.ServerError),
			})
		}
	}
	return &redisLimiter{client: client, cfg: cfg, logger: logger}, nil
}

func (l *redisLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path, l.cfg.ExcludePaths) {
			c.Next()
			return
		}

		key := l.keyForRequest(c)
		if key == "" {
			c.Next()
			return
		}

		allowed, retryAfter, err := l.allow(c.Request.Context(), key)
		if err != nil {
			if l.cfg.FailOpen {
				if l.logger != nil {
					l.logger.Warn("rate limit failed open", zap.Error(err))
				}
				c.Next()
				return
			}

			c.AbortWithStatusJSON(httpStatusServiceUnavailable, code.Failure{
				Code:    code.ServerError,
				Message: code.Text(code.ServerError),
			})
			return
		}

		if !allowed {
			if retryAfter > 0 {
				c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))
			}
			c.AbortWithStatusJSON(httpStatusTooManyRequests, code.Failure{
				Code:    code.RateLimited,
				Message: code.Text(code.RateLimited),
			})
			return
		}

		c.Next()
	}
}

func (l *redisLimiter) keyForRequest(c *gin.Context) string {
	strategy := strings.ToLower(strings.TrimSpace(l.cfg.KeyStrategy))
	token := bearerToken(c.GetHeader("Authorization"))
	clientIP := clientIPFromContext(c)

	switch strategy {
	case "token":
		if token != "" {
			return "token:" + hashToken(token)
		}
		return "ip:" + clientIP
	case "ip":
		return "ip:" + clientIP
	default:
		if token != "" {
			return "token:" + hashToken(token)
		}
		return "ip:" + clientIP
	}
}

func (l *redisLimiter) allow(ctx context.Context, key string) (bool, time.Duration, error) {
	if l.cfg.Limit <= 0 || l.cfg.Window <= 0 {
		return true, 0, nil
	}

	now := time.Now().Unix()
	windowSec := int64(l.cfg.Window.Seconds())
	bucket := now / windowSec
	redisKey := fmt.Sprintf("%s:%s:%d", l.cfg.Prefix, key, bucket)

	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, 0, err
	}

	ttl := time.Duration((bucket+1)*windowSec-now) * time.Second
	if ttl < 0 {
		ttl = 0
	}
	if count == 1 {
		_ = l.client.Expire(ctx, redisKey, ttl+time.Second).Err()
	}

	if count > l.cfg.Limit {
		return false, ttl, nil
	}

	return true, 0, nil
}

func clientIPFromRequest(req *http.Request) string {
	if req == nil {
		return "unknown"
	}
	if forwarded := strings.TrimSpace(req.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(req.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if strings.TrimSpace(req.RemoteAddr) != "" {
		return strings.TrimSpace(req.RemoteAddr)
	}
	return "unknown"
}

func clientIPFromContext(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip != "" {
		return ip
	}
	return clientIPFromRequest(c.Request)
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	h := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func shouldSkipPath(path string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return false
	}
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func configFromEnv() Config {
	cfg := Config{
		Enabled:      envBool("RATE_LIMIT_ENABLED"),
		Limit:        envInt64("RATE_LIMIT_LIMIT", 600),
		Window:       time.Duration(envInt64("RATE_LIMIT_WINDOW_SECONDS", 60)) * time.Second,
		Prefix:       envString("RATE_LIMIT_PREFIX", "rl"),
		FailOpen:     envBoolDefault("RATE_LIMIT_FAIL_OPEN", false),
		KeyStrategy:  envString("RATE_LIMIT_KEY", "auto"),
		ExcludePaths: envStringList("RATE_LIMIT_EXCLUDE_PATHS"),
	}
	if len(cfg.ExcludePaths) == 0 {
		cfg.ExcludePaths = []string{"/system/health"}
	}
	return cfg
}

const (
	httpStatusTooManyRequests    = 429
	httpStatusServiceUnavailable = 503
)

func envBool(key string) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return false
	}
	switch strings.ToLower(val) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func envBoolDefault(key string, def bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return envBool(key)
}

func envInt64(key string, def int64) int64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func envString(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}

func envStringList(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
