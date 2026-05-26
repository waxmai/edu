package interceptor

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/jwtoken"
	"edu-schedule-system/internal/proposal"
)

func (i *interceptor) JWTokenAuthVerify(ctx core.Context) (sessionUserInfo proposal.SessionUserInfo, err core.BusinessError) {
	// 具体 Header 参数，可根据实际情况调整
	headerAuthorizationString := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if headerAuthorizationString == "" {
		err = core.Error(
			http.StatusUnauthorized,
			code.AuthMissingError,
			code.Text(code.AuthMissingError))

		return
	}

	// 验证 JWT 是否合法
	if strings.HasPrefix(strings.ToLower(headerAuthorizationString), "bearer ") {
		headerAuthorizationString = strings.TrimSpace(headerAuthorizationString[7:])
	}

	if headerAuthorizationString == "" {
		err = core.Error(
			http.StatusUnauthorized,
			code.AuthMissingError,
			code.Text(code.AuthMissingError))
		return
	}

	if i.tokenRevoker != nil {
		revoked, revokeErr := i.tokenRevoker.IsRevoked(ctx.RequestContext(), headerAuthorizationString)
		if revokeErr != nil {
			err = core.Error(
				http.StatusUnauthorized,
				code.JWTAuthVerifyError,
				fmt.Sprintf("jwt token 验证失败： %s", revokeErr.Error()))
			return
		}
		if revoked {
			err = core.Error(
				http.StatusUnauthorized,
				code.JWTAuthVerifyError,
				"jwt token 已失效，请重新登录")
			return
		}
	}

	jwtCfg := configs.Get().JWT
	jwtClaims, jwtErr := jwtoken.New(
		jwtCfg.Secret,
		jwtoken.WithIssuer(jwtCfg.Issuer),
		jwtoken.WithAudience(jwtCfg.Audience),
		jwtoken.WithLeeway(time.Duration(jwtCfg.LeewaySeconds)*time.Second),
	).Parse(headerAuthorizationString)
	if jwtErr != nil {
		err = core.Error(
			http.StatusUnauthorized,
			code.JWTAuthVerifyError,
			fmt.Sprintf("jwt token 验证失败： %s", jwtErr.Error()))

		return
	}

	sessionUserInfo = jwtClaims.SessionUserInfo
	if sessionUserInfo.Id <= 0 {
		err = core.Error(
			http.StatusUnauthorized,
			code.JWTAuthVerifyError,
			"jwt token 无效：缺少用户信息")
		return
	}
	if !sessionUserInfo.IsEnabled() {
		err = core.Error(
			http.StatusForbidden,
			code.Forbidden,
			"账号不可用")
		return
	}

	return
}
