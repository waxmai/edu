-- rollback-policy: forward-only
-- rollback-note: Seed data is idempotent; adjust with a new bootstrap/update migration.
-- bootstrap-env-required: BOOTSTRAP_*_PASSWORD or BOOTSTRAP_*_PASSWORD_HASH
-- Demo-only bootstrap users. Do not replace phone/email/real_name with real personal data in source control.
INSERT INTO `sys_user` (`id`, `username`, `password_hash`, `role_code`, `organization_id`, `campus_id`, `real_name`, `phone`, `email`, `status`, `must_change_password`, `failed_login_count`, `locked_until`, `last_password_changed_at`, `remark`) VALUES
(101, 'platform_admin', '{{BOOTSTRAP_PLATFORM_ADMIN_PASSWORD_HASH}}', 'platform_admin', NULL, NULL, 'Demo Platform Admin', '10000000000', 'platform_admin@example.local', 'enabled', 1, 0, NULL, NULL, 'bootstrap local platform admin, rotate bootstrap password immediately before shared testing or deployment'),
(102, 'org_admin', '{{BOOTSTRAP_ORG_ADMIN_PASSWORD_HASH}}', 'org_admin', 1, NULL, 'Demo Org Admin', '10000000001', 'org_admin@example.local', 'enabled', 1, 0, NULL, NULL, 'bootstrap local org admin for organization #1, rotate bootstrap password immediately before shared testing or deployment'),
(103, 'campus_admin', '{{BOOTSTRAP_CAMPUS_ADMIN_PASSWORD_HASH}}', 'campus_admin', 1, 1, 'Demo Campus Admin', '10000000002', 'campus_admin@example.local', 'enabled', 1, 0, NULL, NULL, 'bootstrap local campus admin for campus #1, rotate bootstrap password immediately before shared testing or deployment'),
(104, 'teacher01', '{{BOOTSTRAP_TEACHER_PASSWORD_HASH}}', 'teacher', 1, 1, 'Demo Teacher', '10000000003', 'teacher01@example.local', 'enabled', 1, 0, NULL, NULL, 'bootstrap local teacher for campus #1, rotate bootstrap password immediately before shared testing or deployment')
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  role_code = VALUES(role_code),
  organization_id = VALUES(organization_id),
  campus_id = VALUES(campus_id),
  real_name = VALUES(real_name),
  phone = VALUES(phone),
  email = VALUES(email),
  status = VALUES(status),
  must_change_password = VALUES(must_change_password),
  failed_login_count = VALUES(failed_login_count),
  locked_until = VALUES(locked_until),
  last_password_changed_at = VALUES(last_password_changed_at),
  remark = VALUES(remark);
