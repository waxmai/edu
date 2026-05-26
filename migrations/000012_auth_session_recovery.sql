-- rollback-policy: forward-only
-- rollback-note: Add device session governance and password recovery chain support.

ALTER TABLE `sys_user`
  ADD COLUMN `email` varchar(120) NOT NULL DEFAULT '' COMMENT '邮箱地址，用于账户恢复与二次验证' AFTER `phone`;

CREATE TABLE IF NOT EXISTS `auth_session` (
  `id` varchar(64) NOT NULL COMMENT '会话ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
  `device_id` varchar(64) NOT NULL DEFAULT '' COMMENT '设备指纹ID',
  `device_name` varchar(120) NOT NULL DEFAULT '' COMMENT '设备名称',
  `client_ip` varchar(64) NOT NULL DEFAULT '' COMMENT '登录/刷新IP',
  `user_agent` varchar(255) NOT NULL DEFAULT '' COMMENT '客户端UA',
  `user_agent_hash` varchar(64) NOT NULL DEFAULT '' COMMENT 'UA哈希',
  `access_token_id` varchar(128) NOT NULL DEFAULT '' COMMENT '当前 access token id',
  `refresh_token_id` varchar(128) NOT NULL DEFAULT '' COMMENT '当前 refresh token id',
  `refresh_token_hash` varchar(64) NOT NULL DEFAULT '' COMMENT 'refresh token 摘要',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT 'active/revoked',
  `is_suspicious` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否异常会话',
  `risk_flags` json DEFAULT NULL COMMENT '风险标签',
  `last_seen_at` timestamp NULL DEFAULT NULL COMMENT '最近活跃时间',
  `last_refreshed_at` timestamp NULL DEFAULT NULL COMMENT '最近刷新时间',
  `revoked_at` timestamp NULL DEFAULT NULL COMMENT '撤销时间',
  `revoked_reason` varchar(64) NOT NULL DEFAULT '' COMMENT '撤销原因',
  `recovery_verified_at` timestamp NULL DEFAULT NULL COMMENT '因恢复流程被校验的时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_auth_session_user_status` (`user_id`, `status`),
  KEY `idx_auth_session_device_id` (`device_id`),
  KEY `idx_auth_session_refresh_token_id` (`refresh_token_id`),
  CONSTRAINT `fk_auth_session_user` FOREIGN KEY (`user_id`) REFERENCES `sys_user`(`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证会话表';

CREATE TABLE IF NOT EXISTS `auth_recovery_challenge` (
  `id` varchar(64) NOT NULL COMMENT '恢复挑战ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
  `username` varchar(50) NOT NULL DEFAULT '' COMMENT '用户名快照',
  `channel` varchar(20) NOT NULL COMMENT '主验证通道 email/sms',
  `channel_target` varchar(120) NOT NULL DEFAULT '' COMMENT '主验证目标',
  `code_hash` varchar(64) NOT NULL COMMENT '主验证码摘要',
  `second_channel` varchar(20) NOT NULL DEFAULT '' COMMENT '二次验证通道',
  `second_channel_target` varchar(120) NOT NULL DEFAULT '' COMMENT '二次验证目标',
  `second_code_hash` varchar(64) NOT NULL DEFAULT '' COMMENT '二次验证码摘要',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/verified/consumed/expired',
  `risk_flags` json DEFAULT NULL COMMENT '恢复风险标签',
  `requested_ip` varchar(64) NOT NULL DEFAULT '' COMMENT '请求IP',
  `requested_user_agent` varchar(255) NOT NULL DEFAULT '' COMMENT '请求UA',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `verified_at` timestamp NULL DEFAULT NULL COMMENT '验证完成时间',
  `consumed_at` timestamp NULL DEFAULT NULL COMMENT '消费时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_auth_recovery_user_status` (`user_id`, `status`),
  KEY `idx_auth_recovery_expires_at` (`expires_at`),
  CONSTRAINT `fk_auth_recovery_user` FOREIGN KEY (`user_id`) REFERENCES `sys_user`(`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户恢复挑战表';

CREATE TABLE IF NOT EXISTS `auth_recovery_audit` (
  `id` varchar(64) NOT NULL COMMENT '审计ID',
  `challenge_id` varchar(64) NOT NULL COMMENT '恢复挑战ID',
  `user_id` int(11) unsigned NOT NULL COMMENT '用户ID',
  `action` varchar(64) NOT NULL COMMENT '动作',
  `status` varchar(20) NOT NULL DEFAULT '' COMMENT '状态',
  `detail` text COMMENT '详情',
  `actor_ip` varchar(64) NOT NULL DEFAULT '' COMMENT '操作者IP',
  `user_agent` varchar(255) NOT NULL DEFAULT '' COMMENT '操作者UA',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_auth_recovery_audit_challenge` (`challenge_id`),
  KEY `idx_auth_recovery_audit_user` (`user_id`),
  CONSTRAINT `fk_auth_recovery_audit_challenge` FOREIGN KEY (`challenge_id`) REFERENCES `auth_recovery_challenge`(`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_auth_recovery_audit_user` FOREIGN KEY (`user_id`) REFERENCES `sys_user`(`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户恢复审计表';
