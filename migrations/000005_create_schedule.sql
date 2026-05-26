-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS `schedule` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `student_id` int(11) unsigned NOT NULL COMMENT '学员ID',
  `course_id` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '课程ID',
  `teacher_id` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '教师ID',
  `lesson_package_id` int(11) unsigned DEFAULT NULL COMMENT '关联课时包ID',
  `class_date` date NOT NULL COMMENT '上课日期',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `classroom` varchar(100) NOT NULL DEFAULT '' COMMENT '教室/地点',
  `schedule_status` varchar(20) NOT NULL DEFAULT 'scheduled' COMMENT '状态：scheduled/completed/leave/rescheduled/cancelled/makeup_pending',
  `original_schedule_id` int(11) unsigned DEFAULT NULL COMMENT '原始课程ID',
  `is_makeup` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否补课：0否1是',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_schedule_teacher_time` (`teacher_id`, `class_date`, `start_time`, `end_time`),
  KEY `idx_schedule_student_time` (`student_id`, `class_date`, `start_time`, `end_time`),
  KEY `idx_schedule_status` (`schedule_status`),
  KEY `idx_schedule_original_schedule_id` (`original_schedule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排课表';
