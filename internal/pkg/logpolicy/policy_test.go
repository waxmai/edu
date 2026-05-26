package logpolicy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyRotatesAndTrimsLogFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alerts.ndjson")
	content := strings.Repeat("{\"x\":1}\n", 80)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	result, err := Apply(RetentionPolicy{LogDir: dir, Pattern: "alerts.ndjson", MaxBytes: 128, RotationCount: 3})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if result.Archived == "" {
		t.Fatal("Apply() archived path empty, want rotated file")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if int64(len(body)) > 128 {
		t.Fatalf("trimmed size = %d, want <= 128", len(body))
	}
}
