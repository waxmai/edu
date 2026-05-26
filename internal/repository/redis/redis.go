package redis

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"sync"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/timeutil"
	"edu-schedule-system/internal/pkg/trace"

	redisV8 "github.com/go-redis/redis/v8"
)

var (
	client  *redisV8.Client
	once    sync.Once
	initErr error
)

type loggingHook struct {
	ts time.Time
}

type traceCarrier interface {
	TraceValue() trace.T
}

func (h *loggingHook) BeforeProcess(ctx context.Context, cmd redisV8.Cmder) (context.Context, error) {
	h.ts = time.Now()
	return ctx, nil
}

func (h *loggingHook) AfterProcess(ctx context.Context, cmd redisV8.Cmder) error {
	stdCtx, ok := ctx.(traceCarrier)
	if !ok {
		return nil
	}

	if requestTrace := stdCtx.TraceValue(); requestTrace != nil {
		redisInfo := new(trace.Redis)
		redisInfo.Time = timeutil.CSTLayoutString()
		redisInfo.Stack = fileWithLineNum()
		redisInfo.Cmd = cmd.String()
		redisInfo.CostSeconds = time.Since(h.ts).Seconds()

		requestTrace.AppendRedis(redisInfo)
	}

	return nil
}

func (h *loggingHook) BeforeProcessPipeline(ctx context.Context, cmds []redisV8.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h *loggingHook) AfterProcessPipeline(ctx context.Context, cmds []redisV8.Cmder) error {
	return nil
}

func GetRedisClient() (*redisV8.Client, error) {
	if !configs.Get().Redis.Enabled {
		return nil, errors.New("redis is disabled")
	}

	once.Do(func() {
		cfg := configs.Get().Redis
		client = redisV8.NewClient(&redisV8.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Pass,
			DB:           cfg.Db,
			MaxRetries:   3,
			DialTimeout:  time.Duration(cfg.Pool.DialTimeoutSeconds) * time.Second,
			ReadTimeout:  time.Duration(cfg.Pool.ReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(cfg.Pool.WriteTimeoutSeconds) * time.Second,
			PoolSize:     cfg.Pool.PoolSize,
			MinIdleConns: cfg.Pool.MinIdleConns,
			PoolTimeout:  time.Duration(cfg.Pool.PoolTimeoutSeconds) * time.Second,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			initErr = err
			return
		}

		client.AddHook(&loggingHook{})

	})

	return client, initErr
}

func fileWithLineNum() string {
	_, file, line, ok := runtime.Caller(5)
	if ok {
		return file + ":" + strconv.FormatInt(int64(line), 10)
	}

	return ""
}
