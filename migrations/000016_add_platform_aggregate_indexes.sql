-- rollback-policy: forward-only
-- rollback-note: Add composite indexes for platform aggregate and campus risk queries.

ALTER TABLE `subscription`
  ADD INDEX `idx_subscription_org_status_ends_at` (`organization_id`, `status`, `ends_at`);

ALTER TABLE `sys_user`
  ADD INDEX `idx_sys_user_campus_status` (`campus_id`, `status`);

ALTER TABLE `lesson_package`
  ADD INDEX `idx_lesson_package_campus_student` (`campus_id`, `student_id`);
