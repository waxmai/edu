package core

import (
	"sync"
)

const defaultTraceBodyBytes int64 = 4096

var (
	traceCfgOnce       sync.Once
	traceLogBodyEnable bool
	traceBodyMaxValue  int64
)

func traceBodyEnabled() bool {
	traceCfgOnce.Do(loadTraceConfig)
	return traceLogBodyEnable
}

func traceBodyMaxBytes() int64 {
	traceCfgOnce.Do(loadTraceConfig)
	return traceBodyMaxValue
}

func loadTraceConfig() {
	traceLogBodyEnable = envBool("TRACE_LOG_BODY")
	traceBodyMaxValue = envInt64("TRACE_BODY_MAX_BYTES", defaultTraceBodyBytes)
	if traceBodyMaxValue < 0 {
		traceBodyMaxValue = defaultTraceBodyBytes
	}
}
