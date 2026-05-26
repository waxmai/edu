-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS `student` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `student_name` varchar(50) NOT NULL DEFAULT '' COMMENT '学员姓名',
  `gender` varchar(10) NOT NULL DEFAULT 'unknown' COMMENT '性别：male/female/unknown',
  `grade` varchar(30) NOT NULL DEFAULT '' COMMENT '年级',
  `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '学员手机号',
  `parent_name` varchar(50) NOT NULL DEFAULT '' COMMENT '家长姓名',
  `parent_phone` varchar(20) NOT NULL DEFAULT '' COMMENT '家长电话',
  `subject` varchar(50) NOT NULL DEFAULT '' COMMENT '报读科目',
  `teaching_type` varchar(20) NOT NULL DEFAULT 'one_to_one' COMMENT '授课形式：one_to_one/one_to_many/small_class',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态：active/suspended/graduated',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_student_name` (`student_name`),
  KEY `idx_student_subject` (`subject`),
  KEY `idx_student_status` (`status`),
  KEY `idx_student_parent_phone` (`parent_phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学员表';
