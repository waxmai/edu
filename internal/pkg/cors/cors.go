package cors

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	ginCors "github.com/rs/cors/wrapper/gin"
)

func New() gin.HandlerFunc {
	origins := allowedOrigins()
	if len(origins) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	allowCredentials := shouldAllowCredentials(origins)
	return ginCors.New(ginCors.Options{
		// AllowedOrigins 允许的来源（域名），可以是一个具体的域名或使用通配符 "*" 表示允许所有来源。
		// []string{"http://example.com", "http://another-domain.com"},
		AllowedOrigins: origins,

		// AllowedMethods 允许的 HTTP 方法，例如 "GET"、"POST"、"PUT" 等。
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},

		// AllowedHeaders 允许的请求标头，例如 "Authorization"、"Content-Type" 等。
		AllowedHeaders: []string{"*"},

		// MaxAge 预检请求的最大缓存时间（秒），用于减少预检请求的频率。
		MaxAge: 86400,

		// AllowCredentials 是否允许携带身份凭证（如 Cookie）。
		AllowCredentials: allowCredentials,

		// OptionsPassthrough 是否将 OPTIONS 请求传递给下一个处理函数，设置为 true 可以在 Gin 的路由中使用 OPTIONS 请求处理函数。
		OptionsPassthrough: true,
	})
}

func allowedOrigins() []string {
	envVal := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if envVal == "" {
		return nil
	}

	parts := strings.Split(envVal, ",")
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

func shouldAllowCredentials(origins []string) bool {
	allow := strings.EqualFold(strings.TrimSpace(os.Getenv("CORS_ALLOW_CREDENTIALS")), "true")
	if !allow {
		return false
	}

	for _, o := range origins {
		if o == "*" {
			return false
		}
	}
	return true
}
