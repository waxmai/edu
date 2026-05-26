package internal_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestImportBoundaries(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}|{{join .Imports \",\"}}", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list deps failed: %v\n%s", err, string(out))
	}

	var violations []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		pkg := parts[0]
		imports := ""
		if len(parts) > 1 {
			imports = parts[1]
		}
		importList := strings.Split(imports, ",")

		if strings.HasPrefix(pkg, "edu-schedule-system/internal/pkg/") {
			for _, imp := range importList {
				imp = strings.TrimSpace(imp)
				if strings.HasPrefix(imp, "edu-schedule-system/internal/service") || strings.HasPrefix(imp, "edu-schedule-system/internal/api") || strings.HasPrefix(imp, "edu-schedule-system/internal/router") {
					violations = append(violations, fmt.Sprintf("%s must not depend on business/web layer package %s", pkg, imp))
				}
			}
		}

		if strings.HasPrefix(pkg, "edu-schedule-system/internal/service/") {
			for _, imp := range importList {
				imp = strings.TrimSpace(imp)
				if strings.HasPrefix(imp, "edu-schedule-system/internal/pkg/core") || strings.Contains(imp, "gin-gonic/gin") {
					violations = append(violations, fmt.Sprintf("%s must not depend on web runtime package %s", pkg, imp))
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("import boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}
