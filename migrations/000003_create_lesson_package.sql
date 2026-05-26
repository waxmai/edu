-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS `lesson_package` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `student_id` int(11) unsigned NOT NULL COMMENT '学员ID',
  `course_id` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '课程ID',
  `total_lessons` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '总课时',
  `used_lessons` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '已用课时',
  `remain_lessons` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '剩余课时',
  `total_amount` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '总金额',
  `paid_amount` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '已缴金额',
  `start_date` date DEFAULT NULL COMMENT '生效日期',
  `end_date` date DEFAULT NULL COMMENT '截止日期',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/active/exhausted/expired/closed',
  `low_lesson_threshold` decimal(10,2) NOT NULL DEFAULT 3.00 COMMENT '低课时提醒阈值',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_lesson_package_student_id` (`student_id`),
  KEY `idx_lesson_package_status` (`status`),
  KEY `idx_lesson_package_end_date` (`end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课时包表';
