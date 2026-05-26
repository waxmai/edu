package interceptor

import (
	"net/http"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/core"
)

func (i *interceptor) RequirePasswordChanged() core.HandlerFunc {
	return func(ctx core.Context) {
		user := ctx.SessionUserInfo()
		if user.Id <= 0 {
			ctx.AbortWithError(core.Error(http.StatusUnauthorized, code.AuthMissingError, code.Text(code.AuthMissingError)))
			return
		}
		if user.MustChangePassword {
			ctx.AbortWithError(core.Error(http.StatusForbidden, code.Forbidden, "首次登录后请先修改密码"))
			return
		}
	}
}
