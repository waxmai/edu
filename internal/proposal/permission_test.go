package proposal

import "testing"

func TestPlatformPresetRolePermissionBoundaries(t *testing.T) {
	ops := BuildAccessProfile(RolePlatformOps)
	if !containsPermission(ops.Permissions, PermPlatformTenantList) {
		t.Fatalf("platform_ops should list tenants")
	}
	if containsPermission(ops.Permissions, PermPlatformSubscriptionUpdate) {
		t.Fatalf("platform_ops must not update subscriptions")
	}
	if containsPermission(ops.Permissions, PermPlatformAuditExport) {
		t.Fatalf("platform_ops must not export audit logs")
	}

	finance := BuildAccessProfile(RolePlatformFinance)
	if !containsPermission(finance.Permissions, PermPlatformSubscriptionUpdate) {
		t.Fatalf("platform_finance should update subscriptions")
	}
	if containsPermission(finance.Permissions, PermPlatformAuditExport) {
		t.Fatalf("platform_finance must not export audit logs")
	}
	if containsPermission(finance.Permissions, PermPlatformTenantUpdate) {
		t.Fatalf("platform_finance must not update tenants")
	}

	auditor := BuildAccessProfile(RolePlatformAuditor)
	if !containsPermission(auditor.Permissions, PermPlatformAuditList) || !containsPermission(auditor.Permissions, PermPlatformAuditExport) {
		t.Fatalf("platform_auditor should list and export audit logs")
	}
	if containsPermission(auditor.Permissions, PermPlatformTenantUpdate) || containsPermission(auditor.Permissions, PermPlatformSubscriptionUpdate) {
		t.Fatalf("platform_auditor must not mutate tenants or subscriptions")
	}

	support := BuildAccessProfile(RolePlatformSupport)
	if !containsPermission(support.Permissions, PermPlatformRecoveryAlertResolve) {
		t.Fatalf("platform_support should resolve recovery alerts")
	}
	if containsPermission(support.Permissions, PermPaymentUpdate) || containsPermission(support.Permissions, PermPlatformSubscriptionUpdate) {
		t.Fatalf("platform_support must not mutate payment or subscription data")
	}
}

func containsPermission(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
