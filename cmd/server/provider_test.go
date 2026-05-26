package server

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"
)

type closeDBRepo struct {
	wErr error
	rErr error
}

func (r *closeDBRepo) GetDbR() *gorm.DB           { return nil }
func (r *closeDBRepo) GetDbW() *gorm.DB           { return nil }
func (r *closeDBRepo) DbRClose() error            { return r.rErr }
func (r *closeDBRepo) DbWClose() error            { return r.wErr }
func (r *closeDBRepo) Ping(context.Context) error { return nil }
func (r *closeDBRepo) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type closeCacheRepo struct{ err error }

func (r *closeCacheRepo) Set(context.Context, string, interface{}, time.Duration) error { return nil }
func (r *closeCacheRepo) Get(context.Context, string) (string, error)                   { return "", nil }
func (r *closeCacheRepo) Del(context.Context, string) error                             { return nil }
func (r *closeCacheRepo) Close() error                                                  { return r.err }
func (r *closeCacheRepo) Ping(context.Context) error                                    { return nil }
func (r *closeCacheRepo) Exists(context.Context, string) (bool, error)                  { return false, nil }
func (r *closeCacheRepo) GetOrSet(context.Context, string, time.Duration, func() (interface{}, error)) (interface{}, error) {
	return nil, nil
}

func TestAppCloseAggregatesResourceErrors(t *testing.T) {
	dbErr := errors.New("db close")
	cacheErr := errors.New("cache close")
	app := &App{
		Server: &http.Server{},
		DB:     &closeDBRepo{wErr: dbErr},
		Cache:  &closeCacheRepo{err: cacheErr},
	}

	err := app.Close(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("Close() error = %v, want db error", err)
	}
	if !errors.Is(err, cacheErr) {
		t.Fatalf("Close() error = %v, want cache error", err)
	}
}

func TestAppCloseNilSafe(t *testing.T) {
	var app *App
	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("nil Close() error = %v, want nil", err)
	}
}
