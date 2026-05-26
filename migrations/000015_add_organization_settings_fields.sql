-- rollback-policy: forward-only
-- rollback-note: Add tenant settings fields for branding, notification, and security policy metadata.

ALTER TABLE `organization`
  ADD COLUMN `brand_name` varchar(100) NOT NULL DEFAULT '' COMMENT '品牌名称' AFTER `timezone`,
  ADD COLUMN `notification_email` varchar(100) NOT NULL DEFAULT '' COMMENT '通知邮箱' AFTER `brand_name`,
  ADD COLUMN `security_policy` json DEFAULT NULL COMMENT '安全策略配置' AFTER `notification_email`;
