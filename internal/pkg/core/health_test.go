package core

import (
	stdctx "context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestHealthRouteAlwaysRegistered(t *testing.T) {
	mux, err := New(zap.NewNop())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/system/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("/system/health status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func TestReadyRouteRunsConfiguredChecks(t *testing.T) {
	mux, err := New(zap.NewNop(), WithReadinessChecks(map[string]ReadinessCheck{
		"ok": func(stdctx.Context) error { return nil },
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/system/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("/system/ready status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func TestReadyRouteReturnsServiceUnavailableOnFailedCheck(t *testing.T) {
	mux, err := New(zap.NewNop(), WithReadinessChecks(map[string]ReadinessCheck{
		"db": func(stdctx.Context) error { return errors.New("down") },
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/system/ready", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("/system/ready status = %d, want %d; body=%s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
}
