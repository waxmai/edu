package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type bootstrapSecretSpec struct {
	placeholder string
	hashEnv     string
	plainEnv    string
}

func bootstrapSecretSpecs() []bootstrapSecretSpec {
	return []bootstrapSecretSpec{
		{placeholder: "{{BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH}}", hashEnv: "BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH", plainEnv: "BOOTSTRAP_PLATFORM_ADMIN_PASSWORD"},
		{placeholder: "{{BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH}}", hashEnv: "BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH", plainEnv: "BOOTSTRAP_ORG_ADMIN_PASSWORD"},
		{placeholder: "{{BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH}}", hashEnv: "BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH", plainEnv: "BOOTSTRAP_CAMPUS_ADMIN_PASSWORD"},
		{placeholder: "{{BOOTSTRAP_TEACHER_PASSWORD_HASH}}", hashEnv: "BOOTSTRAP_TEACHER_PASSWORD_HASH", plainEnv: "BOOTSTRAP_TEACHER_PASSWORD"},
	}
}

func bootstrapEnvReplacements() map[string]string {
	result := make(map[string]string, len(bootstrapSecretSpecs()))
	for _, spec := range bootstrapSecretSpecs() {
		result[spec.placeholder] = resolveBootstrapSecretValue(spec)
	}
	return result
}

func resolveBootstrapSecretValue(spec bootstrapSecretSpec) string {
	hashValue := strings.TrimSpace(os.Getenv(spec.hashEnv))
	if hashValue != "" {
		return hashValue
	}
	plainValue := strings.TrimSpace(os.Getenv(spec.plainEnv))
	if plainValue == "" {
		return ""
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainValue), 10)
	if err != nil {
		panic(fmt.Sprintf("generate bootstrap bcrypt for %s: %v", spec.plainEnv, err))
	}
	return string(hashed)
}

func validateBootstrapSecrets() error {
	for _, spec := range bootstrapSecretSpecs() {
		value := resolveBootstrapSecretValue(spec)
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("bootstrap requires env value for %s or %s", spec.hashEnv, spec.plainEnv)
		}
	}
	return nil
}

func applyBootstrapReplacements(sqlText string) string {
	result := sqlText
	for placeholder, value := range bootstrapEnvReplacements() {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
