-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
ALTER TABLE `sys_user`
  ADD COLUMN `failed_login_count` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '连续登录失败次数' AFTER `status`,
  ADD COLUMN `locked_until` timestamp NULL DEFAULT NULL COMMENT '锁定截止时间' AFTER `failed_login_count`;

UPDATE `sys_user`
SET `failed_login_count` = 0
WHERE `failed_login_count` IS NULL;
