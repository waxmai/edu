package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"edu-schedule-system/cmd/server"
	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/shutdown"

	"go.uber.org/zap"
)

// @title Edu Schedule System 接口文档
// @version v0.0.1

// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 初始化环境配置
	env.Init()
	if err := configs.Init(); err != nil {
		log.Fatalf("config init err: %v", err)
	}

	// 依赖注入初始化应用
	app, cleanup, err := server.InitializeApp()
	if err != nil {
		log.Fatalf("app init err: %v", err)
	}
	// 确保在程序退出时释放资源
	defer cleanup()

	go func() {
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("http server startup err", zap.Error(err))
		}
	}()

	// 优雅关闭
	shutdown.Close(
		func() {
			shutdownTimeout := time.Duration(configs.Get().Server.ShutdownTimeoutSeconds) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()

			if err := app.Close(ctx); err != nil {
				app.Logger.Error("app close err", zap.Error(err))
			}
		},
	)
}
