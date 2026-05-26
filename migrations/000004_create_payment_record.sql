-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.
CREATE TABLE IF NOT EXISTS `payment_record` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `student_id` int(11) unsigned NOT NULL COMMENT '学员ID',
  `lesson_package_id` int(11) unsigned NOT NULL COMMENT '课时包ID',
  `payment_type` varchar(20) NOT NULL DEFAULT 'signup' COMMENT '缴费类型：signup/renewal',
  `amount` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '缴费金额',
  `payment_method` varchar(20) NOT NULL DEFAULT 'other' COMMENT '支付方式：cash/wechat/alipay/bank_transfer/other',
  `payment_time` datetime NOT NULL COMMENT '缴费时间',
  `payment_status` varchar(20) NOT NULL DEFAULT 'paid' COMMENT '状态：paid/pending/partial/refunded',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_payment_record_student_id` (`student_id`),
  KEY `idx_payment_record_package_id` (`lesson_package_id`),
  KEY `idx_payment_record_time` (`payment_time`),
  KEY `idx_payment_record_status` (`payment_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='缴费记录表';
