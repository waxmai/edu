package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestQueryServiceListFiltersAuditRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	content := "{\"time\":\"2026-05-15T10:00:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"auth.login\",\"module\":\"auth\",\"target_id\":1,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"t1\",\"detail\":{\"username\":\"admin\"}}\n" +
		"{\"time\":\"2026-05-15T10:01:00+08:00\",\"level\":\"info\",\"msg\":\"trace-log\",\"path\":\"/api/v1/auth/login\"}\n" +
		"{\"time\":\"2026-05-15T10:02:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"user.create\",\"module\":\"user\",\"target_id\":3,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"t2\"}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := NewQueryService(path)
	resp, err := service.List(Query{Module: "auth", Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("List() total/items = %d/%d, want 1/1", resp.Total, len(resp.Items))
	}
	if resp.Items[0].Action != "auth.login" {
		t.Fatalf("List() action = %q, want auth.login", resp.Items[0].Action)
	}
}

func TestQueryServiceListSupportsTraceTimeAndPagination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	content := "{\"time\":\"2026-05-15T10:00:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"auth.login\",\"module\":\"auth\",\"target_id\":1,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"trace-a\"}\n" +
		"{\"time\":\"2026-05-15T10:05:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"user.create\",\"module\":\"user\",\"target_id\":3,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"trace-b\"}\n" +
		"{\"time\":\"2026-05-15T10:10:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"user.update\",\"module\":\"user\",\"target_id\":4,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"trace-c\"}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := NewQueryService(path)
	resp, err := service.List(Query{StartTime: "2026-05-15T10:04:00+08:00", EndTime: "2026-05-15T10:10:00+08:00", Offset: 0, Limit: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("List() total = %d, want 2", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].TraceID != "trace-c" {
		t.Fatalf("List() first trace = %q, want trace-c", resp.Items[0].TraceID)
	}
	resp, err = service.List(Query{TraceID: "trace-b", Limit: 10})
	if err != nil {
		t.Fatalf("List() trace filter error = %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].TraceID != "trace-b" {
		t.Fatalf("List() trace filter mismatch: %#v", resp)
	}
	meta := service.Meta()
	if len(meta["actions"]) != 3 || len(meta["modules"]) != 2 {
		t.Fatalf("Meta() = %#v, want 3 actions and 2 modules", meta)
	}
}

func TestQueryServiceExportReturnsRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	content := "{\"time\":\"2026-05-15T10:00:00+08:00\",\"level\":\"info\",\"msg\":\"audit-event\",\"category\":\"audit\",\"action\":\"auth.login\",\"module\":\"auth\",\"target_id\":1,\"actor_id\":2,\"actor_username\":\"admin\",\"actor_role\":\"platform_admin\",\"trace_id\":\"trace-a\",\"detail\":{\"username\":\"admin\"}}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := NewQueryService(path)
	rows, err := service.Export(Query{Limit: 100})
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Export() rows = %d, want 1", len(rows))
	}
	if rows[0].Action != "auth.login" || rows[0].Detail == "" {
		t.Fatalf("Export() row = %#v", rows[0])
	}
}
