-- rollback-policy: forward-only
-- rollback-note: Add notification delivery records for SaaS notification observability.

CREATE TABLE IF NOT EXISTS `notification_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `organization_id` int unsigned NOT NULL DEFAULT 0 COMMENT '机构ID，平台级或未知为0',
  `channel` varchar(20) NOT NULL COMMENT '通知通道：email/sms',
  `scene` varchar(64) NOT NULL COMMENT '通知场景',
  `template_code` varchar(100) NOT NULL DEFAULT '' COMMENT '模板编码',
  `target_masked` varchar(120) NOT NULL DEFAULT '' COMMENT '脱敏后的目标地址',
  `provider` varchar(64) NOT NULL DEFAULT '' COMMENT '通知供应商',
  `status` varchar(20) NOT NULL COMMENT '发送状态：accepted/error',
  `error_code` varchar(64) NOT NULL DEFAULT '' COMMENT '归一化错误码',
  `error_message` varchar(512) NOT NULL DEFAULT '' COMMENT '错误信息，不包含敏感目标或验证码',
  `request_id` varchar(128) NOT NULL DEFAULT '' COMMENT '供应商请求ID或幂等键',
  `attempts` int unsigned NOT NULL DEFAULT 1 COMMENT '尝试次数',
  `sent_at` timestamp NULL DEFAULT NULL COMMENT '发送完成时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_notification_record_org_created` (`organization_id`, `created_at`),
  KEY `idx_notification_record_scene_status_created` (`scene`, `status`, `created_at`),
  KEY `idx_notification_record_channel_status_created` (`channel`, `status`, `created_at`),
  KEY `idx_notification_record_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知发送记录表';
