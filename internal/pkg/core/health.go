package core

import (
	stdctx "context"
	"net/http"
	"time"

	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/timeutil"
)

func registerHealthRoutes(mux *mux, opt *option) {
	system := mux.Group("/system")
	{
		// 健康检查
		system.GET("/health", func(ctx Context) {
			resp := &struct {
				Time        string `json:"time"`
				Environment string `json:"environment"`
				Host        string `json:"host"`
				Status      string `json:"status"`
			}{
				Time:        timeutil.CSTLayoutString(),
				Environment: env.Active().Value(),
				Host:        ctx.Host(),
				Status:      "ok",
			}
			ctx.Payload(resp)
		})

		// 就绪检查：用于容器和负载均衡确认依赖是否可用。
		system.GET("/ready", func(ctx Context) {
			checks := opt.readinessChecks
			statuses := make(map[string]string, len(checks))
			ready := true

			checkCtx, cancel := stdctx.WithTimeout(ctx.Request().Context(), 2*time.Second)
			defer cancel()

			for name, check := range checks {
				if check == nil {
					statuses[name] = "ok"
					continue
				}
				if err := check(checkCtx); err != nil {
					ready = false
					statuses[name] = err.Error()
					if opt.alertRules != nil {
						opt.alertRules.RecordReadyCheck(checkCtx, name, err)
					}
					continue
				}
				statuses[name] = "ok"
				if opt.alertRules != nil {
					opt.alertRules.RecordReadyCheck(checkCtx, name, nil)
				}
			}

			resp := &struct {
				Time        string            `json:"time"`
				Environment string            `json:"environment"`
				Host        string            `json:"host"`
				Status      string            `json:"status"`
				Checks      map[string]string `json:"checks"`
			}{
				Time:        timeutil.CSTLayoutString(),
				Environment: env.Active().Value(),
				Host:        ctx.Host(),
				Status:      "ok",
				Checks:      statuses,
			}

			if !ready {
				resp.Status = "unready"
				ctx.PayloadWithStatus(http.StatusServiceUnavailable, resp)
				return
			}

			ctx.Payload(resp)
		})
	}
}
