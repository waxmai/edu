-- rollback-policy: forward-only
-- rollback-note: Revert schema changes with a new forward migration.

UPDATE `lesson_package`
SET `used_lessons` = LEAST(GREATEST(`used_lessons`, 0), `total_lessons`),
    `remain_lessons` = GREATEST(`total_lessons` - LEAST(GREATEST(`used_lessons`, 0), `total_lessons`), 0),
    `paid_amount` = LEAST(GREATEST(`paid_amount`, 0), `total_amount`)
WHERE `paid_amount` > `total_amount`
   OR `used_lessons` < 0
   OR `used_lessons` > `total_lessons`
   OR `remain_lessons` < 0
   OR `remain_lessons` > `total_lessons`
   OR ABS((`used_lessons` + `remain_lessons`) - `total_lessons`) >= 0.001;

UPDATE `schedule` s
LEFT JOIN `schedule` os ON os.`id` = s.`original_schedule_id`
SET s.`original_schedule_id` = NULL
WHERE s.`original_schedule_id` IS NOT NULL
  AND os.`id` IS NULL;

UPDATE `reschedule_record` r
LEFT JOIN `sys_user` u ON u.`id` = r.`operator_id`
SET r.`operator_id` = NULL
WHERE r.`operator_id` IS NOT NULL
  AND u.`id` IS NULL;

SET @sql = IF (
  EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'fk_lesson_package_student'
  ),
  'SELECT 1',
  'ALTER TABLE `lesson_package` ADD CONSTRAINT `fk_lesson_package_student` FOREIGN KEY (`student_id`) REFERENCES `student`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF (
  EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_total_lessons_non_negative'
  ),
  'SELECT 1',
  'ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_total_lessons_non_negative` CHECK (`total_lessons` >= 0)'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_used_lessons_non_negative'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_used_lessons_non_negative` CHECK (`used_lessons` >= 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_remain_lessons_non_negative'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_remain_lessons_non_negative` CHECK (`remain_lessons` >= 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_total_amount_non_negative'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_total_amount_non_negative` CHECK (`total_amount` >= 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_paid_amount_non_negative'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_paid_amount_non_negative` CHECK (`paid_amount` >= 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_low_lesson_threshold_positive'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_low_lesson_threshold_positive` CHECK (`low_lesson_threshold` > 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_used_not_exceed_total'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_used_not_exceed_total` CHECK (`used_lessons` <= `total_lessons`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_remain_not_exceed_total'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_remain_not_exceed_total` CHECK (`remain_lessons` <= `total_lessons`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_lessons_balance'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_lessons_balance` CHECK (ABS((`used_lessons` + `remain_lessons`) - `total_lessons`) < 0.001)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_paid_not_exceed_total'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_paid_not_exceed_total` CHECK (`paid_amount` <= `total_amount`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_status'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_status` CHECK (`status` IN (''pending'',''active'',''exhausted'',''expired'',''closed''))');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'lesson_package' AND constraint_name = 'chk_lesson_package_date_range'),'SELECT 1','ALTER TABLE `lesson_package` ADD CONSTRAINT `chk_lesson_package_date_range` CHECK (`end_date` IS NULL OR `start_date` IS NULL OR `end_date` >= `start_date`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'payment_record' AND constraint_name = 'fk_payment_record_student'),'SELECT 1','ALTER TABLE `payment_record` ADD CONSTRAINT `fk_payment_record_student` FOREIGN KEY (`student_id`) REFERENCES `student`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'payment_record' AND constraint_name = 'fk_payment_record_lesson_package'),'SELECT 1','ALTER TABLE `payment_record` ADD CONSTRAINT `fk_payment_record_lesson_package` FOREIGN KEY (`lesson_package_id`) REFERENCES `lesson_package`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'payment_record' AND constraint_name = 'chk_payment_record_amount_positive'),'SELECT 1','ALTER TABLE `payment_record` ADD CONSTRAINT `chk_payment_record_amount_positive` CHECK (`amount` > 0)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'payment_record' AND constraint_name = 'chk_payment_record_type'),'SELECT 1','ALTER TABLE `payment_record` ADD CONSTRAINT `chk_payment_record_type` CHECK (`payment_type` IN (''signup'',''renewal''))');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'payment_record' AND constraint_name = 'chk_payment_record_status'),'SELECT 1','ALTER TABLE `payment_record` ADD CONSTRAINT `chk_payment_record_status` CHECK (`payment_status` IN (''paid'',''pending'',''partial'',''refunded''))');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'fk_schedule_student'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `fk_schedule_student` FOREIGN KEY (`student_id`) REFERENCES `student`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'fk_schedule_lesson_package'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `fk_schedule_lesson_package` FOREIGN KEY (`lesson_package_id`) REFERENCES `lesson_package`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'fk_schedule_original_schedule'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `fk_schedule_original_schedule` FOREIGN KEY (`original_schedule_id`) REFERENCES `schedule`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'chk_schedule_status'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `chk_schedule_status` CHECK (`schedule_status` IN (''scheduled'',''completed'',''leave'',''rescheduled'',''cancelled'',''makeup_pending''))');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'chk_schedule_time_range'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `chk_schedule_time_range` CHECK (`end_time` > `start_time`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'uk_schedule_teacher_exact_slot'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `uk_schedule_teacher_exact_slot` UNIQUE (`teacher_id`, `start_time`, `end_time`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'schedule' AND constraint_name = 'uk_schedule_student_exact_slot'),'SELECT 1','ALTER TABLE `schedule` ADD CONSTRAINT `uk_schedule_student_exact_slot` UNIQUE (`student_id`, `start_time`, `end_time`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'fk_reschedule_record_old_schedule'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `fk_reschedule_record_old_schedule` FOREIGN KEY (`old_schedule_id`) REFERENCES `schedule`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'fk_reschedule_record_new_schedule'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `fk_reschedule_record_new_schedule` FOREIGN KEY (`new_schedule_id`) REFERENCES `schedule`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'fk_reschedule_record_operator'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `fk_reschedule_record_operator` FOREIGN KEY (`operator_id`) REFERENCES `sys_user`(`id`) ON UPDATE RESTRICT ON DELETE RESTRICT');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'chk_reschedule_record_operation_type'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `chk_reschedule_record_operation_type` CHECK (`operation_type` IN (''leave'',''reschedule'',''makeup'',''cancel''))');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'chk_reschedule_record_old_new_not_same'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `chk_reschedule_record_old_new_not_same` CHECK (`new_schedule_id` IS NULL OR `new_schedule_id` <> `old_schedule_id`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql = IF (EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'reschedule_record' AND constraint_name = 'uk_reschedule_record_old_operation_new'),'SELECT 1','ALTER TABLE `reschedule_record` ADD CONSTRAINT `uk_reschedule_record_old_operation_new` UNIQUE (`old_schedule_id`, `operation_type`, `new_schedule_id`)');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
