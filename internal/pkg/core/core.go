package core

import (
	"fmt"

	"edu-schedule-system/internal/pkg/color"
	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// see https://patorjk.com/software/taag/#p=testall&f=Graffiti&t=edu-schedule-system
const _UI = `edu-schedule-system success!`

// New logger required
func New(logger *zap.Logger, options ...Option) (Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	if env.Active().IsDev() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	mux := &mux{
		engine: gin.New(),
	}

	fmt.Println(color.Blue(_UI))

	configureTrustedProxies(mux.engine, logger)
	registerBaseMiddleware(mux.engine, logger)

	withoutTracePaths := defaultWithoutTracePaths()

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	registerFeatureEndpoints(mux.engine, opt)
	mux.engine.Use(recoveryMiddleware(logger))
	mux.engine.Use(requestLifecycleMiddleware(logger, opt, withoutTracePaths))

	mux.engine.NoMethod(wrapHandlers(DisableTraceLog)...)
	mux.engine.NoRoute(wrapHandlers(DisableTraceLog)...)

	registerHealthRoutes(mux, opt)

	return mux, nil
}
