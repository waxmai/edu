package redis

import (
	"context"
	"errors"
	"time"

	"edu-schedule-system/configs"

	"github.com/go-redis/redis/v8"
)

const nilPlaceholder = "__redis_nil__"

// Repo 定义 Redis 仓库接口
type Repo interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Close() error
	Ping(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)
	// CacheAside 模式 helper
	GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (interface{}, error)) (interface{}, error)
}

type cacheRepo struct {
	client *redis.Client
}

type noopRepo struct{}

func NewCache() (Repo, error) {
	if !configs.Get().Redis.Enabled {
		return noopRepo{}, nil
	}

	client, err := GetRedisClient()
	if err != nil {
		return nil, err
	}
	return &cacheRepo{
		client: client,
	}, nil
}

func (c *cacheRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *cacheRepo) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *cacheRepo) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *cacheRepo) Close() error {
	return c.client.Close()
}

func (c *cacheRepo) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *cacheRepo) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetOrSet Cache-Aside 实现
func (c *cacheRepo) GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (interface{}, error)) (interface{}, error) {
	// 1. Try get from cache
	val, err := c.Get(ctx, key)
	if err == nil {
		if val == nilPlaceholder {
			return nil, nil
		}
		return val, nil
	}
	if err != redis.Nil {
		return nil, err // Redis error
	}

	// 2. Cache miss, fetch from DB
	data, err := fetch()
	if err != nil {
		return nil, err
	}

	// 3. Set cache
	cacheValue := data
	if data == nil {
		cacheValue = nilPlaceholder
	}
	if err := c.Set(ctx, key, cacheValue, expiration); err != nil {
		return data, nil
	}

	return data, nil
}

func (noopRepo) Set(context.Context, string, interface{}, time.Duration) error {
	return nil
}

func (noopRepo) Get(context.Context, string) (string, error) {
	return "", redis.Nil
}

func (noopRepo) Del(context.Context, string) error {
	return nil
}

func (noopRepo) Close() error {
	return nil
}

func (noopRepo) Ping(context.Context) error {
	return nil
}

func (noopRepo) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (noopRepo) GetOrSet(ctx context.Context, key string, expiration time.Duration, fetch func() (interface{}, error)) (interface{}, error) {
	return fetch()
}

func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
