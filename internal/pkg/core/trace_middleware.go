package core

import (
	"net/http"
	"net/url"
	"time"

	"edu-schedule-system/internal/pkg/trace"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func defaultWithoutTracePaths() map[string]bool {
	return map[string]bool{
		"/metrics": true,

		"/debug/pprof/":             true,
		"/debug/pprof/cmdline":      true,
		"/debug/pprof/profile":      true,
		"/debug/pprof/symbol":       true,
		"/debug/pprof/trace":        true,
		"/debug/pprof/allocs":       true,
		"/debug/pprof/block":        true,
		"/debug/pprof/goroutine":    true,
		"/debug/pprof/heap":         true,
		"/debug/pprof/mutex":        true,
		"/debug/pprof/threadcreate": true,

		"/favicon.ico": true,

		"/system/health": true,
	}
}

func initRequestTrace(context Context, path string, withoutTracePaths map[string]bool) {
	if withoutTracePaths[path] {
		return
	}
	if traceID := context.GetHeader(trace.Header); traceID != "" {
		context.setTrace(trace.New(traceID))
		return
	}
	context.setTrace(trace.New(""))
}

func writeTraceLog(logger *zap.Logger, ctx *gin.Context, context Context, ts time.Time, response interface{}, businessCode int, businessCodeMsg string, abortErr error) {
	var t *trace.Trace
	if x := context.Trace(); x != nil {
		t = x.(*trace.Trace)
	} else {
		return
	}

	decodedURL, _ := url.QueryUnescape(ctx.Request.URL.RequestURI())

	traceHeader := map[string]string{
		"Content-Type": ctx.GetHeader("Content-Type"),
	}

	var rawBody []byte
	if traceBodyEnabled() {
		rawBody = context.RawData()
		if maxBytes := traceBodyMaxBytes(); maxBytes > 0 && int64(len(rawBody)) > maxBytes {
			rawBody = rawBody[:maxBytes]
		}
	}

	t.WithRequest(&trace.Request{
		TTL:        "un-limit",
		Method:     ctx.Request.Method,
		DecodedURL: decodedURL,
		Header:     traceHeader,
		Body:       string(rawBody),
	})

	var responseBody interface{}
	if traceBodyEnabled() && response != nil {
		responseBody = response
	}

	t.WithResponse(&trace.Response{
		Header:          ctx.Writer.Header(),
		HttpCode:        ctx.Writer.Status(),
		HttpCodeMsg:     http.StatusText(ctx.Writer.Status()),
		BusinessCode:    businessCode,
		BusinessCodeMsg: businessCodeMsg,
		Body:            responseBody,
		CostSeconds:     time.Since(ts).Seconds(),
	})

	t.Success = !ctx.IsAborted() && (ctx.Writer.Status() == http.StatusOK)
	t.CostSeconds = time.Since(ts).Seconds()

	logger.Info("trace-log",
		zap.Any("method", ctx.Request.Method),
		zap.Any("path", decodedURL),
		zap.Any("http_code", ctx.Writer.Status()),
		zap.Any("business_code", businessCode),
		zap.Any("success", t.Success),
		zap.Any("cost_seconds", t.CostSeconds),
		zap.Any("trace_id", t.Identifier),
		zap.Any("trace_info", t),
		zap.Error(abortErr),
	)
}
