package interceptor

import "edu-schedule-system/internal/pkg/core"

func (i *interceptor) Authenticate() core.HandlerFunc {
	return core.WrapAuthHandler(i.JWTokenAuthVerify)
}
