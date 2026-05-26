package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestParseRollbackPolicy(t *testing.T) {
	policy := parseRollbackPolicy("-- rollback-policy: manual\nSELECT 1;")
	if policy != "manual" {
		t.Fatalf("parseRollbackPolicy() = %q, want manual", policy)
	}

	policy = parseRollbackPolicy("SELECT 1;")
	if policy != "" {
		t.Fatalf("parseRollbackPolicy() = %q, want empty string", policy)
	}

	policy = parseRollbackPolicy("-- rollback-policy: forward-only\nSELECT 1;")
	if policy != "forward-only" {
		t.Fatalf("parseRollbackPolicy() bootstrap = %q, want forward-only", policy)
	}
}

func TestMigrationChecksumEncoding(t *testing.T) {
	sum := sha256.Sum256([]byte("hello"))
	got := hex.EncodeToString(sum[:])
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("checksum = %q, want %q", got, want)
	}
}

func TestValidateBootstrapSecrets(t *testing.T) {
	hashKeys := []string{
		"BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH",
		"BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH",
		"BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH",
		"BOOTSTRAP_TEACHER_PASSWORD_HASH",
	}
	plainKeys := []string{
		"BOOTSTRAP_PLATFORM_ADMIN_PASSWORD",
		"BOOTSTRAP_ORG_ADMIN_PASSWORD",
		"BOOTSTRAP_CAMPUS_ADMIN_PASSWORD",
		"BOOTSTRAP_TEACHER_PASSWORD",
	}
	for _, key := range append(hashKeys, plainKeys...) {
		t.Setenv(key, "")
	}
	if err := validateBootstrapSecrets(); err == nil {
		t.Fatal("validateBootstrapSecrets() error = nil, want missing env failure")
	}

	for _, key := range hashKeys {
		t.Setenv(key, "$2a$10$exampleexampleexampleexampleexampleexampleexampleexample")
	}
	if err := validateBootstrapSecrets(); err != nil {
		t.Fatalf("validateBootstrapSecrets() hash env error = %v, want nil", err)
	}
}

func TestValidateBootstrapSecretsAllowsPlainPasswords(t *testing.T) {
	keys := []string{
		"BOOTSTRAP_PLATFORM_ADMIN_PASSWORD",
		"BOOTSTRAP_ORG_ADMIN_PASSWORD",
		"BOOTSTRAP_CAMPUS_ADMIN_PASSWORD",
		"BOOTSTRAP_TEACHER_PASSWORD",
	}
	for _, key := range keys {
		t.Setenv(key, "StrongPass123!")
	}
	if err := validateBootstrapSecrets(); err != nil {
		t.Fatalf("validateBootstrapSecrets() plain env error = %v, want nil", err)
	}
}

func TestApplyBootstrapReplacements(t *testing.T) {
	t.Setenv("BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH", "hash-platform")
	t.Setenv("BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH", "hash-org")
	t.Setenv("BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH", "hash-campus")
	t.Setenv("BOOTSTRAP_TEACHER_PASSWORD_HASH", "hash-teacher")

	input := strings.Join([]string{
		"{{BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH}}",
		"{{BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH}}",
		"{{BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH}}",
		"{{BOOTSTRAP_TEACHER_PASSWORD_HASH}}",
	}, ",")
	got := applyBootstrapReplacements(input)
	want := "hash-platform,hash-org,hash-campus,hash-teacher"
	if got != want {
		t.Fatalf("applyBootstrapReplacements() = %q, want %q", got, want)
	}
}

func TestApplyBootstrapReplacementsUsesPlainPasswords(t *testing.T) {
	t.Setenv("BOOTSTRAP_PLATFORM_ADMIN_PASSWORD", "PlatformInit123!")
	t.Setenv("BOOTSTRAP_ORG_ADMIN_PASSWORD", "OrgInit123!")
	t.Setenv("BOOTSTRAP_CAMPUS_ADMIN_PASSWORD", "CampusInit123!")
	t.Setenv("BOOTSTRAP_TEACHER_PASSWORD", "TeacherInit123!")

	got := applyBootstrapReplacements("{{BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH}}")
	if !strings.HasPrefix(got, "$2") {
		t.Fatalf("applyBootstrapReplacements() = %q, want bcrypt hash", got)
	}
}
