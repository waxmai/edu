package core

import (
	_ "edu-schedule-system/docs"
	"edu-schedule-system/internal/pkg/cors"
	"edu-schedule-system/internal/pkg/env"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerFeatureEndpoints(engine *gin.Engine, opt *option) {
	if opt.enablePProf && envBool("PPROF_ENABLED") && !env.Active().IsPro() {
		pprof.Register(engine) // register pprof to gin
	}

	if opt.enableSwagger && envBool("SWAGGER_ENABLED") && !env.Active().IsPro() {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // register swagger
	}

	if opt.enablePrometheus && envBool("METRICS_ENABLED") {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler())) // register prometheus
	}

	if opt.enableCors {
		engine.Use(cors.New())
	}
}
