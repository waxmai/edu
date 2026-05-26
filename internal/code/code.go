package code

import (
	_ "embed"

	"edu-schedule-system/configs"
)

//go:embed code.go
var ByteCodeFile []byte

// Failure 错误时返回结构
type Failure struct {
	Code    int    `json:"code"`    // 业务码
	Message string `json:"message"` // 描述信息
}

// Response 成功时返回结构
type Response struct {
	Code    int         `json:"code"`     // 业务码
	Message string      `json:"message"`  // 描述信息
	Data    interface{} `json:"data"`     // 数据
	TraceID string      `json:"trace_id"` // TraceID
}

const (
	ServerError        = 10101
	ParamBindError     = 10102
	JWTAuthVerifyError = 10103
	AuthMissingError   = 10104
	RecordNotFound     = 10105
	RateLimited        = 10106
	PayloadTooLarge    = 10107
	Conflict           = 10108
	Forbidden          = 10109
	DependencyFailed   = 10110

	// Add more business codes here
)

func Text(code int) string {
	lang := configs.Get().Language.Local

	if lang == configs.ZhCN {
		return zhCNText[code]
	}

	if lang == configs.EnUS {
		return enUSText[code]
	}

	return zhCNText[code]
}

var zhCNText = map[int]string{
	ServerError:        "服务器内部错误",
	ParamBindError:     "参数绑定错误",
	JWTAuthVerifyError: "JWT 验证失败",
	AuthMissingError:   "缺少 Authorization 信息",
	RecordNotFound:     "记录不存在",
	RateLimited:        "请求过频",
	PayloadTooLarge:    "请求体过大",
	Conflict:           "资源冲突",
	Forbidden:          "无权访问",
	DependencyFailed:   "依赖服务不可用",
}

var enUSText = map[int]string{
	ServerError:        "Internal Server Error",
	ParamBindError:     "Parameter Binding Error",
	JWTAuthVerifyError: "JWT Validation Failed",
	AuthMissingError:   "Missing Authorization",
	RecordNotFound:     "Record Not Found",
	RateLimited:        "Too Many Requests",
	PayloadTooLarge:    "Request Entity Too Large",
	Conflict:           "Conflict",
	Forbidden:          "Forbidden",
	DependencyFailed:   "Dependency Service Unavailable",
}
