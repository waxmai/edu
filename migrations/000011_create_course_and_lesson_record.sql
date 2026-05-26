-- rollback-policy: forward-only
-- rollback-note: Add formal course and lesson_record support.

CREATE TABLE IF NOT EXISTS `course` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `course_name` varchar(100) NOT NULL COMMENT '课程名称',
  `subject` varchar(50) NOT NULL COMMENT '科目',
  `course_type` varchar(30) NOT NULL DEFAULT '' COMMENT '课程类型',
  `duration_minutes` int NOT NULL DEFAULT 60 COMMENT '标准时长，单位分钟',
  `fee_standard` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '标准收费',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `status` varchar(20) NOT NULL DEFAULT 'enabled' COMMENT '状态：enabled/disabled',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_course_subject` (`subject`),
  KEY `idx_course_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程定义表';

CREATE TABLE IF NOT EXISTS `lesson_record` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `schedule_id` int unsigned NOT NULL COMMENT '排课ID',
  `student_id` int unsigned NOT NULL COMMENT '学员ID',
  `teacher_id` int unsigned NOT NULL COMMENT '教师ID',
  `attendance_status` varchar(20) NOT NULL COMMENT '出勤状态：present/late/absent/leave',
  `lesson_content` text COMMENT '上课内容',
  `homework` text COMMENT '作业',
  `feedback` text COMMENT '反馈',
  `lesson_deducted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已扣课时：0否1是',
  `deduct_lesson_count` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '本次扣减课时数',
  `recorded_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_lesson_record_schedule_id` (`schedule_id`),
  KEY `idx_lesson_record_student_id` (`student_id`),
  KEY `idx_lesson_record_teacher_id` (`teacher_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='上课记录表';
