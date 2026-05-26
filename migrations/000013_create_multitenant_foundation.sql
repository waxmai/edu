-- rollback-policy: forward-only
-- rollback-note: Introduce SaaS multitenant organization/campus/subscription foundation and migrate legacy single-campus data.

CREATE TABLE IF NOT EXISTS `organization` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `org_code` varchar(50) NOT NULL COMMENT '机构编码',
  `org_name` varchar(100) NOT NULL COMMENT '机构名称',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态：active/inactive',
  `subscription_status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '订阅状态：trial/active/past_due/suspended/expired',
  `edition_code` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '版本编码',
  `feature_flags` json DEFAULT NULL COMMENT '机构级功能开关',
  `timezone` varchar(50) NOT NULL DEFAULT 'Asia/Shanghai' COMMENT '机构时区',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_organization_org_code` (`org_code`),
  KEY `idx_organization_status` (`status`),
  KEY `idx_organization_subscription_status` (`subscription_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机构表';

CREATE TABLE IF NOT EXISTS `campus` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `organization_id` int unsigned NOT NULL COMMENT '机构ID',
  `campus_code` varchar(50) NOT NULL COMMENT '校区编码',
  `campus_name` varchar(100) NOT NULL COMMENT '校区名称',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态：active/inactive',
  `feature_flags` json DEFAULT NULL COMMENT '校区级功能开关',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_campus_org_code` (`organization_id`, `campus_code`),
  KEY `idx_campus_org_status` (`organization_id`, `status`),
  CONSTRAINT `fk_campus_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='校区表';

CREATE TABLE IF NOT EXISTS `subscription` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `organization_id` int unsigned NOT NULL COMMENT '机构ID',
  `plan_code` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '套餐编码',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '订阅状态：trial/active/past_due/suspended/expired',
  `starts_at` datetime DEFAULT NULL COMMENT '生效时间',
  `ends_at` datetime DEFAULT NULL COMMENT '到期时间',
  `feature_flags` json DEFAULT NULL COMMENT '订阅功能开关',
  `max_campuses` int unsigned NOT NULL DEFAULT 1 COMMENT '最大校区数',
  `max_users` int unsigned NOT NULL DEFAULT 100 COMMENT '最大用户数',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_subscription_org` (`organization_id`),
  KEY `idx_subscription_status` (`status`),
  CONSTRAINT `fk_subscription_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机构订阅表';

INSERT INTO `organization` (`id`, `org_code`, `org_name`, `status`, `subscription_status`, `edition_code`, `feature_flags`, `timezone`, `remark`)
VALUES (
  1,
  'default-org',
  '默认机构',
  'active',
  'active',
  'standard',
  JSON_ARRAY('auth','user_management','student','course','lesson_package','payment','schedule','lesson_record','reschedule','platform_management'),
  'Asia/Shanghai',
  'legacy single-tenant migration anchor'
)
ON DUPLICATE KEY UPDATE
  `org_name` = VALUES(`org_name`),
  `status` = VALUES(`status`),
  `subscription_status` = VALUES(`subscription_status`),
  `edition_code` = VALUES(`edition_code`),
  `feature_flags` = VALUES(`feature_flags`),
  `timezone` = VALUES(`timezone`),
  `remark` = VALUES(`remark`);

INSERT INTO `campus` (`id`, `organization_id`, `campus_code`, `campus_name`, `status`, `feature_flags`, `remark`)
VALUES (
  1,
  1,
  'main-campus',
  '默认校区',
  'active',
  JSON_ARRAY('student','course','lesson_package','payment','schedule','lesson_record','reschedule'),
  'legacy single-campus migration anchor'
)
ON DUPLICATE KEY UPDATE
  `campus_name` = VALUES(`campus_name`),
  `status` = VALUES(`status`),
  `feature_flags` = VALUES(`feature_flags`),
  `remark` = VALUES(`remark`);

INSERT INTO `subscription` (`id`, `organization_id`, `plan_code`, `status`, `starts_at`, `ends_at`, `feature_flags`, `max_campuses`, `max_users`, `remark`)
VALUES (
  1,
  1,
  'standard',
  'active',
  NOW(),
  NULL,
  JSON_ARRAY('auth','user_management','student','course','lesson_package','payment','schedule','lesson_record','reschedule','platform_management'),
  10,
  500,
  'legacy single-tenant migration subscription'
)
ON DUPLICATE KEY UPDATE
  `plan_code` = VALUES(`plan_code`),
  `status` = VALUES(`status`),
  `starts_at` = VALUES(`starts_at`),
  `ends_at` = VALUES(`ends_at`),
  `feature_flags` = VALUES(`feature_flags`),
  `max_campuses` = VALUES(`max_campuses`),
  `max_users` = VALUES(`max_users`),
  `remark` = VALUES(`remark`);

ALTER TABLE `sys_user`
  MODIFY COLUMN `role_code` varchar(32) NOT NULL COMMENT '角色编码：platform_admin/org_admin/campus_admin/teacher',
  ADD COLUMN `organization_id` int unsigned DEFAULT NULL COMMENT '所属机构ID' AFTER `role_code`,
  ADD COLUMN `campus_id` int unsigned DEFAULT NULL COMMENT '所属校区ID' AFTER `organization_id`;

UPDATE `sys_user`
SET `role_code` = CASE
    WHEN `role_code` = 'admin' THEN 'platform_admin'
    WHEN `role_code` = 'teacher' THEN 'teacher'
    ELSE `role_code`
  END,
  `organization_id` = CASE
    WHEN `role_code` = 'admin' THEN NULL
    ELSE 1
  END,
  `campus_id` = CASE
    WHEN `role_code` = 'teacher' THEN 1
    ELSE NULL
  END
WHERE `organization_id` IS NULL OR `campus_id` IS NULL OR `role_code` IN ('admin', 'teacher');

ALTER TABLE `auth_session`
  ADD COLUMN `organization_id` int unsigned DEFAULT NULL COMMENT '所属机构ID快照' AFTER `user_id`,
  ADD COLUMN `campus_id` int unsigned DEFAULT NULL COMMENT '所属校区ID快照' AFTER `organization_id`,
  ADD COLUMN `role_code` varchar(32) NOT NULL DEFAULT '' COMMENT '角色编码快照' AFTER `campus_id`;

UPDATE `auth_session` s
JOIN `sys_user` u ON u.`id` = s.`user_id`
SET s.`organization_id` = u.`organization_id`,
    s.`campus_id` = u.`campus_id`,
    s.`role_code` = u.`role_code`
WHERE s.`organization_id` IS NULL OR s.`role_code` = '';

ALTER TABLE `student`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `course`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `lesson_package`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `payment_record`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `schedule`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `lesson_record`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;
ALTER TABLE `reschedule_record`
  ADD COLUMN `organization_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属机构ID' AFTER `id`,
  ADD COLUMN `campus_id` int unsigned NOT NULL DEFAULT 1 COMMENT '所属校区ID' AFTER `organization_id`;

UPDATE `student` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `course` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `lesson_package` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `payment_record` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `schedule` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `lesson_record` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;
UPDATE `reschedule_record` SET `organization_id` = 1, `campus_id` = 1 WHERE `organization_id` <> 1 OR `campus_id` <> 1;

ALTER TABLE `sys_user`
  ADD CONSTRAINT `fk_sys_user_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_sys_user_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `auth_session`
  ADD CONSTRAINT `fk_auth_session_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_auth_session_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `student`
  ADD KEY `idx_student_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_student_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_student_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `course`
  ADD KEY `idx_course_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_course_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_course_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `lesson_package`
  ADD KEY `idx_lesson_package_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_lesson_package_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_lesson_package_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `payment_record`
  ADD KEY `idx_payment_record_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_payment_record_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_payment_record_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `schedule`
  ADD KEY `idx_schedule_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_schedule_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_schedule_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `lesson_record`
  ADD KEY `idx_lesson_record_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_lesson_record_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_lesson_record_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE `reschedule_record`
  ADD KEY `idx_reschedule_record_org_campus` (`organization_id`, `campus_id`),
  ADD CONSTRAINT `fk_reschedule_record_organization` FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_reschedule_record_campus` FOREIGN KEY (`campus_id`) REFERENCES `campus`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT;
