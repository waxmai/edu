-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
ALTER TABLE `sys_user`
  ADD COLUMN `must_change_password` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否首次登录后必须修改密码' AFTER `status`,
  ADD COLUMN `last_password_changed_at` timestamp NULL DEFAULT NULL COMMENT '最近一次主动修改密码时间' AFTER `must_change_password`;

UPDATE `sys_user`
SET `must_change_password` = 1
WHERE `must_change_password` IS NULL OR `must_change_password` = 0;
