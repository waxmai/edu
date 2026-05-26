-- rollback-policy: additive
ALTER TABLE `subscription`
  ADD COLUMN `follow_up_status` varchar(50) NOT NULL DEFAULT '' COMMENT '续费跟进状态' AFTER `remark`,
  ADD COLUMN `follow_up_owner` varchar(100) NOT NULL DEFAULT '' COMMENT '跟进负责人' AFTER `follow_up_status`,
  ADD COLUMN `follow_up_note` varchar(255) NOT NULL DEFAULT '' COMMENT '跟进备注' AFTER `follow_up_owner`,
  ADD COLUMN `last_contact_at` datetime NULL COMMENT '最近联系时间' AFTER `follow_up_note`;
