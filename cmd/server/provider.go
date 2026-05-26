package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/pkg/env"
	"edu-schedule-system/internal/pkg/logger"
	"edu-schedule-system/internal/pkg/timeutil"
	"edu-schedule-system/internal/repository/mysql"
	"edu-schedule-system/internal/repository/redis"

	"go.uber.org/zap"
)

type App struct {
	Server *http.Server
	Logger *zap.Logger
	DB     mysql.Repo
	Cache  redis.Repo
}

func NewApp(server *http.Server, logger *zap.Logger, db mysql.Repo, cache redis.Repo) *App {
	return &App{
		Server: server,
		Logger: logger,
		DB:     db,
		Cache:  cache,
	}
}

// Close centralizes runtime resource cleanup for graceful shutdown.
func (a *App) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}

	var errs []error
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			a.logCloseError("server shutdown err", err)
			errs = append(errs, err)
		}
	}
	if a.DB != nil {
		if err := a.DB.DbWClose(); err != nil {
			a.logCloseError("dbw close err", err)
			errs = append(errs, err)
		}
		if err := a.DB.DbRClose(); err != nil {
			a.logCloseError("dbr close err", err)
			errs = append(errs, err)
		}
	}
	if a.Cache != nil {
		if err := a.Cache.Close(); err != nil {
			a.logCloseError("cache close err", err)
			errs = append(errs, err)
		}
	}
	if a.Logger != nil {
		if err := a.Logger.Sync(); err != nil && !isIgnorableSyncError(err) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *App) logCloseError(message string, err error) {
	if a.Logger != nil {
		a.Logger.Error(message, zap.Error(err))
	}
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid argument") || strings.Contains(msg, "inappropriate ioctl")
}

func NewLogger() (*zap.Logger, error) {
	// 初始化环境配置 (Assume env.Init() is called before or we call it here?
	// env.Init() is currently in main. It's safer to call it in main before Wire.)

	return logger.NewJSONLogger(
		logger.WithOutputInConsole(),
		logger.WithField("domain", fmt.Sprintf("%s[%s]", configs.ProjectName, env.Active().Value())),
		logger.WithTimeLayout(timeutil.CSTLayout),
		logger.WithFileRotationP(configs.ProjectAccessLogFile),
	)
}

func NewHTTPServer(mux core.Mux) *http.Server {
	cfg := configs.Get().Server
	return &http.Server{
		Addr:              cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeoutSeconds) * time.Second,
		ReadTimeout:       time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
	}
}
