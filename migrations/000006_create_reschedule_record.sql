-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS `reschedule_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `old_schedule_id` int(11) unsigned NOT NULL COMMENT '原排课ID',
  `new_schedule_id` int(11) unsigned DEFAULT NULL COMMENT '新排课ID',
  `operation_type` varchar(20) NOT NULL COMMENT '类型：leave/reschedule/makeup/cancel',
  `reason` varchar(255) NOT NULL DEFAULT '' COMMENT '原因',
  `operator_id` int(11) unsigned DEFAULT NULL COMMENT '操作人',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_reschedule_operation_type` (`operation_type`),
  KEY `idx_reschedule_created_at` (`created_at`),
  KEY `idx_reschedule_old_schedule_id` (`old_schedule_id`),
  KEY `idx_reschedule_new_schedule_id` (`new_schedule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调补课记录表';
