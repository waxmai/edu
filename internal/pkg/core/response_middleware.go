package core

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/timeutil"
	"edu-schedule-system/internal/pkg/trace"
	"edu-schedule-system/internal/proposal"

	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func requestLifecycleMiddleware(logger *zap.Logger, opt *option, withoutTracePaths map[string]bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Writer.Status() == http.StatusNotFound {
			return
		}

		ts := time.Now()

		context := newContext(ctx)
		defer releaseContext(context)

		context.init()
		context.setLogger(logger)
		context.ableRecordMetrics()

		initRequestTrace(context, ctx.Request.URL.Path, withoutTracePaths)

		defer func() {
			var (
				response        interface{}
				businessCode    int
				businessCodeMsg string
				abortErr        error
				traceID         string
			)

			if ct := context.Trace(); ct != nil {
				context.SetHeader(trace.Header, ct.ID())
				traceID = ct.ID()
			}

			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))
				context.AbortWithError(Error(
					http.StatusInternalServerError,
					code.ServerError,
					code.Text(code.ServerError)),
				)

				if alertHandler := opt.alertNotify; alertHandler != nil {
					alertHandler(&proposal.AlertMessage{
						ProjectName:  configs.ProjectName,
						Env:          env.Active().Value(),
						TraceID:      traceID,
						HOST:         context.Host(),
						URI:          context.URI(),
						Method:       context.Method(),
						ErrorMessage: err,
						ErrorStack:   stackInfo,
						Time:         time.Now().Format(timeutil.CSTLayout),
					})
				}
			}

			if ctx.IsAborted() {
				for i := range ctx.Errors {
					multierr.AppendInto(&abortErr, ctx.Errors[i])
				}

				if err := context.abortError(); err != nil {
					if err.IsAlert() {
						if alertHandler := opt.alertNotify; alertHandler != nil {
							alertHandler(&proposal.AlertMessage{
								ProjectName:  configs.ProjectName,
								Env:          env.Active().Value(),
								TraceID:      traceID,
								HOST:         context.Host(),
								URI:          context.URI(),
								Method:       context.Method(),
								ErrorMessage: err.Message(),
								ErrorStack:   fmt.Sprintf("%+v", err.StackError()),
								Time:         time.Now().Format(timeutil.CSTLayout),
							})
						}
					}

					multierr.AppendInto(&abortErr, err.StackError())
					businessCode = err.BusinessCode()
					businessCodeMsg = err.Message()
					response = &code.Failure{
						Code:    businessCode,
						Message: businessCodeMsg,
					}
					ctx.JSON(err.HTTPCode(), response)
				}
			}

			response = context.getPayload()
			if response != nil {
				ctx.JSON(context.responseStatus(), response)
			}

			if opt.recordHandler != nil && context.isRecordMetrics() {
				path := context.RoutePath()
				if path == "" {
					path = context.Path()
				}
				if alias := context.Alias(); alias != "" {
					path = alias
				}

				opt.recordHandler(&proposal.MetricsMessage{
					HOST:         context.Host(),
					Path:         path,
					Method:       context.Method(),
					HTTPCode:     ctx.Writer.Status(),
					BusinessCode: businessCode,
					CostSeconds:  time.Since(ts).Seconds(),
					IsSuccess:    !ctx.IsAborted() && (ctx.Writer.Status() == http.StatusOK),
				})
			}

			writeTraceLog(logger, ctx, context, ts, response, businessCode, businessCodeMsg, abortErr)
		}()

		ctx.Next()
	}
}
