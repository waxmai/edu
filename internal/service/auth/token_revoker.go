package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/repository/redis"
)

const logoutTokenTTL = accessTokenTTL + 5*time.Minute

type TokenRevoker interface {
	Revoke(ctx context.Context, token string, expiration time.Duration) error
	IsRevoked(ctx context.Context, token string) (bool, error)
}

type noopTokenRevoker struct{}

type memoryTokenRevoker struct {
	mu      sync.RWMutex
	revoked map[string]time.Time
}

type redisTokenRevoker struct {
	cache redis.Repo
}

func NewTokenRevoker(cache redis.Repo) TokenRevoker {
	if !configs.Get().Redis.Enabled {
		return &memoryTokenRevoker{revoked: map[string]time.Time{}}
	}
	if cache == nil {
		return noopTokenRevoker{}
	}
	return &redisTokenRevoker{cache: cache}
}

func (noopTokenRevoker) Revoke(context.Context, string, time.Duration) error {
	return nil
}

func (noopTokenRevoker) IsRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func (r *memoryTokenRevoker) Revoke(_ context.Context, token string, expiration time.Duration) error {
	if r == nil || token == "" {
		return nil
	}
	if expiration <= 0 {
		expiration = logoutTokenTTL
	}
	key := revokeTokenKey(token)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.revoked == nil {
		r.revoked = map[string]time.Time{}
	}
	r.revoked[key] = time.Now().Add(expiration)
	return nil
}

func (r *memoryTokenRevoker) IsRevoked(_ context.Context, token string) (bool, error) {
	if r == nil || token == "" {
		return false, nil
	}
	key := revokeTokenKey(token)
	now := time.Now()

	r.mu.RLock()
	expiresAt, ok := r.revoked[key]
	r.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if now.Before(expiresAt) {
		return true, nil
	}

	r.mu.Lock()
	if current, exists := r.revoked[key]; exists && !now.Before(current) {
		delete(r.revoked, key)
	}
	r.mu.Unlock()
	return false, nil
}

func (r *redisTokenRevoker) Revoke(ctx context.Context, token string, expiration time.Duration) error {
	if r == nil || r.cache == nil || token == "" {
		return nil
	}
	if expiration <= 0 {
		expiration = logoutTokenTTL
	}
	return r.cache.Set(ctx, revokeTokenKey(token), "1", expiration)
}

func (r *redisTokenRevoker) IsRevoked(ctx context.Context, token string) (bool, error) {
	if r == nil || r.cache == nil || token == "" {
		return false, nil
	}
	ok, err := r.cache.Exists(ctx, revokeTokenKey(token))
	if err == nil {
		return ok, nil
	}
	if redis.IsNil(err) {
		return false, nil
	}

	value, getErr := r.cache.Get(ctx, revokeTokenKey(token))
	if getErr == nil {
		return value != "", nil
	}
	if redis.IsNil(getErr) {
		return false, nil
	}
	return false, errors.Join(err, getErr)
}

func revokeTokenKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("auth:revoked:%s", hex.EncodeToString(sum[:]))
}
