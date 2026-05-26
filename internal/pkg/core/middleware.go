package core

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func configureTrustedProxies(engine *gin.Engine, logger *zap.Logger) {
	if proxies := envStringList("TRUSTED_PROXIES"); len(proxies) > 0 {
		if err := engine.SetTrustedProxies(proxies); err != nil {
			logger.Warn("trusted proxies invalid", zap.Error(err))
		}
	} else {
		if err := engine.SetTrustedProxies(nil); err != nil {
			logger.Warn("trusted proxies disabled", zap.Error(err))
		}
	}
}

func registerBaseMiddleware(engine *gin.Engine, logger *zap.Logger) {
	engine.Use(securityHeadersMiddleware())

	maxBodyBytes := configs.Get().Server.MaxBodyBytes
	if envMax := envInt64("HTTP_MAX_BODY_BYTES", 0); envMax > 0 {
		maxBodyBytes = envMax
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = 1 << 20
	}
	engine.Use(maxBodyBytesMiddleware(maxBodyBytes))

	if limiter := ratelimit.NewFromEnv(logger); limiter != nil {
		engine.Use(limiter)
	}
}

func recoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", string(debug.Stack())))
			}
		}()

		ctx.Next()
	}
}

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-XSS-Protection", "0")
		c.Next()
	}
}

func maxBodyBytesMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes > 0 && c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, code.Failure{
				Code:    code.PayloadTooLarge,
				Message: code.Text(code.PayloadTooLarge),
			})
			return
		}
		if maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
