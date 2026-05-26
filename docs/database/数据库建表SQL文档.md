# 数据库建表 SQL 文档

## 1. 说明

本文档基于《个人教培排课、课时与收费管理系统 PRD》整理，用于指导 MySQL 8.0 数据库建表。

设计原则：
- 尽量贴合 PRD 中的业务闭环
- 统一审计字段，便于追踪数据
- 对核心状态字段使用约束枚举值说明
- 为排课冲突检测、学员检索、课时与收费统计预留索引

字符集建议：`utf8mb4`
存储引擎建议：`InnoDB`

---

## 2. 建库 SQL

```sql
CREATE DATABASE IF NOT EXISTS edu_schedule_system
DEFAULT CHARACTER SET utf8mb4
DEFAULT COLLATE utf8mb4_0900_ai_ci;

USE edu_schedule_system;
```

---

## 3. 建表 SQL

### 3.1 用户表 `sys_user`

```sql
DROP TABLE IF EXISTS sys_user;
CREATE TABLE sys_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    username VARCHAR(50) NOT NULL COMMENT '登录账号',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    role_code VARCHAR(20) NOT NULL COMMENT '角色编码：admin/teacher',
    real_name VARCHAR(50) NOT NULL COMMENT '真实姓名',
    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    status VARCHAR(20) NOT NULL DEFAULT 'enabled' COMMENT '状态：enabled/disabled/locked',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_sys_user_username (username),
    KEY idx_sys_user_role_code (role_code),
    KEY idx_sys_user_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';
```

### 3.2 学员表 `student`

```sql
DROP TABLE IF EXISTS student;
CREATE TABLE student (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    student_name VARCHAR(50) NOT NULL COMMENT '学员姓名',
    gender VARCHAR(10) DEFAULT NULL COMMENT '性别：male/female/unknown',
    grade VARCHAR(30) DEFAULT NULL COMMENT '年级',
    phone VARCHAR(20) DEFAULT NULL COMMENT '学员手机号',
    parent_name VARCHAR(50) DEFAULT NULL COMMENT '家长姓名',
    parent_phone VARCHAR(20) DEFAULT NULL COMMENT '家长电话',
    subject VARCHAR(50) NOT NULL COMMENT '报读科目',
    teaching_type VARCHAR(20) NOT NULL COMMENT '授课形式：one_to_one/one_to_many/small_class',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active/suspended/graduated',
    remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
    created_by BIGINT DEFAULT NULL COMMENT '创建人',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY idx_student_name (student_name),
    KEY idx_student_subject (subject),
    KEY idx_student_status (status),
    KEY idx_student_parent_phone (parent_phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学员表';
```

### 3.3 课程定义表 `course`

```sql
DROP TABLE IF EXISTS course;
CREATE TABLE course (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    course_name VARCHAR(100) NOT NULL COMMENT '课程名称',
    subject VARCHAR(50) NOT NULL COMMENT '科目',
    course_type VARCHAR(30) DEFAULT NULL COMMENT '课程类型',
    duration_minutes INT NOT NULL DEFAULT 60 COMMENT '标准时长，单位分钟',
    fee_standard DECIMAL(10,2) DEFAULT NULL COMMENT '标准收费',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    status VARCHAR(20) NOT NULL DEFAULT 'enabled' COMMENT '状态：enabled/disabled',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY idx_course_subject (subject),
    KEY idx_course_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程定义表';
```

### 3.4 课时包表 `lesson_package`

```sql
DROP TABLE IF EXISTS lesson_package;
CREATE TABLE lesson_package (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    student_id BIGINT NOT NULL COMMENT '学员ID',
    course_id BIGINT NOT NULL COMMENT '课程ID',
    total_lessons DECIMAL(10,2) NOT NULL COMMENT '总课时',
    used_lessons DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '已用课时',
    remain_lessons DECIMAL(10,2) NOT NULL COMMENT '剩余课时',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总金额',
    paid_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '已缴金额',
    start_date DATE DEFAULT NULL COMMENT '生效日期',
    end_date DATE DEFAULT NULL COMMENT '截止日期',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/active/exhausted/expired/closed',
    low_lesson_threshold DECIMAL(10,2) NOT NULL DEFAULT 3 COMMENT '低课时提醒阈值',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT fk_lesson_package_student FOREIGN KEY (student_id) REFERENCES student(id),
    CONSTRAINT fk_lesson_package_course FOREIGN KEY (course_id) REFERENCES course(id),
    KEY idx_lesson_package_student_id (student_id),
    KEY idx_lesson_package_status (status),
    KEY idx_lesson_package_end_date (end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课时包表';
```

### 3.5 排课表 `schedule`

```sql
DROP TABLE IF EXISTS schedule;
CREATE TABLE schedule (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    student_id BIGINT NOT NULL COMMENT '学员ID',
    course_id BIGINT NOT NULL COMMENT '课程ID',
    teacher_id BIGINT NOT NULL COMMENT '教师ID',
    lesson_package_id BIGINT DEFAULT NULL COMMENT '关联课时包ID',
    class_date DATE NOT NULL COMMENT '上课日期',
    start_time DATETIME NOT NULL COMMENT '开始时间',
    end_time DATETIME NOT NULL COMMENT '结束时间',
    classroom VARCHAR(100) DEFAULT NULL COMMENT '教室/地点',
    schedule_status VARCHAR(20) NOT NULL DEFAULT 'scheduled' COMMENT '状态：scheduled/completed/leave/rescheduled/cancelled/makeup_pending',
    original_schedule_id BIGINT DEFAULT NULL COMMENT '原始课程ID，用于调课补课追溯',
    is_makeup TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否补课：0否1是',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT fk_schedule_student FOREIGN KEY (student_id) REFERENCES student(id),
    CONSTRAINT fk_schedule_course FOREIGN KEY (course_id) REFERENCES course(id),
    CONSTRAINT fk_schedule_teacher FOREIGN KEY (teacher_id) REFERENCES sys_user(id),
    CONSTRAINT fk_schedule_package FOREIGN KEY (lesson_package_id) REFERENCES lesson_package(id),
    KEY idx_schedule_teacher_time (teacher_id, class_date, start_time, end_time),
    KEY idx_schedule_student_time (student_id, class_date, start_time, end_time),
    KEY idx_schedule_status (schedule_status),
    KEY idx_schedule_original_schedule_id (original_schedule_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排课表';
```

### 3.6 调补课记录表 `reschedule_record`

```sql
DROP TABLE IF EXISTS reschedule_record;
CREATE TABLE reschedule_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    old_schedule_id BIGINT NOT NULL COMMENT '原排课ID',
    new_schedule_id BIGINT DEFAULT NULL COMMENT '新排课ID',
    operation_type VARCHAR(20) NOT NULL COMMENT '类型：leave/reschedule/makeup/cancel',
    reason VARCHAR(255) DEFAULT NULL COMMENT '原因',
    operator_id BIGINT DEFAULT NULL COMMENT '操作人',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    CONSTRAINT fk_reschedule_old_schedule FOREIGN KEY (old_schedule_id) REFERENCES schedule(id),
    CONSTRAINT fk_reschedule_new_schedule FOREIGN KEY (new_schedule_id) REFERENCES schedule(id),
    CONSTRAINT fk_reschedule_operator FOREIGN KEY (operator_id) REFERENCES sys_user(id),
    KEY idx_reschedule_operation_type (operation_type),
    KEY idx_reschedule_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调补课记录表';
```

### 3.7 上课记录表 `lesson_record`

```sql
DROP TABLE IF EXISTS lesson_record;
CREATE TABLE lesson_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    schedule_id BIGINT NOT NULL COMMENT '排课ID',
    student_id BIGINT NOT NULL COMMENT '学员ID',
    teacher_id BIGINT NOT NULL COMMENT '教师ID',
    attendance_status VARCHAR(20) NOT NULL COMMENT '出勤状态：present/late/absent/leave',
    lesson_content TEXT DEFAULT NULL COMMENT '上课内容',
    homework TEXT DEFAULT NULL COMMENT '作业',
    feedback TEXT DEFAULT NULL COMMENT '反馈',
    lesson_deducted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已扣课时：0否1是',
    deduct_lesson_count DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '本次扣减课时数',
    recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT fk_lesson_record_schedule FOREIGN KEY (schedule_id) REFERENCES schedule(id),
    CONSTRAINT fk_lesson_record_student FOREIGN KEY (student_id) REFERENCES student(id),
    CONSTRAINT fk_lesson_record_teacher FOREIGN KEY (teacher_id) REFERENCES sys_user(id),
    UNIQUE KEY uk_lesson_record_schedule_id (schedule_id),
    KEY idx_lesson_record_student_id (student_id),
    KEY idx_lesson_record_teacher_id (teacher_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='上课记录表';
```

### 3.8 缴费记录表 `payment_record`

```sql
DROP TABLE IF EXISTS payment_record;
CREATE TABLE payment_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    student_id BIGINT NOT NULL COMMENT '学员ID',
    lesson_package_id BIGINT NOT NULL COMMENT '课时包ID',
    payment_type VARCHAR(20) NOT NULL COMMENT '缴费类型：signup/renewal',
    amount DECIMAL(10,2) NOT NULL COMMENT '缴费金额',
    payment_method VARCHAR(20) NOT NULL COMMENT '支付方式：cash/wechat/alipay/bank_transfer/other',
    payment_time DATETIME NOT NULL COMMENT '缴费时间',
    payment_status VARCHAR(20) NOT NULL DEFAULT 'paid' COMMENT '状态：paid/pending/partial/refunded',
    remark VARCHAR(255) DEFAULT NULL COMMENT '备注',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT fk_payment_record_student FOREIGN KEY (student_id) REFERENCES student(id),
    CONSTRAINT fk_payment_record_package FOREIGN KEY (lesson_package_id) REFERENCES lesson_package(id),
    KEY idx_payment_record_student_id (student_id),
    KEY idx_payment_record_package_id (lesson_package_id),
    KEY idx_payment_record_time (payment_time),
    KEY idx_payment_record_status (payment_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='缴费记录表';
```

### 3.9 操作日志表 `operation_log`

```sql
DROP TABLE IF EXISTS operation_log;
CREATE TABLE operation_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    module_name VARCHAR(50) NOT NULL COMMENT '模块名',
    business_id BIGINT DEFAULT NULL COMMENT '业务ID',
    operation_type VARCHAR(30) NOT NULL COMMENT '操作类型',
    operation_desc VARCHAR(255) DEFAULT NULL COMMENT '操作描述',
    operator_id BIGINT DEFAULT NULL COMMENT '操作人ID',
    operator_name VARCHAR(50) DEFAULT NULL COMMENT '操作人姓名',
    request_data JSON DEFAULT NULL COMMENT '请求数据',
    response_data JSON DEFAULT NULL COMMENT '响应数据',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY idx_operation_log_module_name (module_name),
    KEY idx_operation_log_business_id (business_id),
    KEY idx_operation_log_operator_id (operator_id),
    KEY idx_operation_log_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';
```

---

## 4. 初始化基础数据 SQL

```sql
INSERT INTO sys_user (username, password_hash, role_code, real_name, phone, status)
VALUES
('admin', '$2a$10$replace_with_real_bcrypt_hash', 'admin', '系统管理员', '10000000000', 'enabled');

INSERT INTO course (course_name, subject, course_type, duration_minutes, fee_standard, status)
VALUES
('初中数学一对一', '数学', 'one_to_one', 120, 300.00, 'enabled'),
('高中英语一对一', '英语', 'one_to_one', 120, 320.00, 'enabled');
```

---

## 5. 关键约束与说明

### 5.1 排课冲突校验
数据库层面可通过索引加速查询，但冲突判断建议在业务层实现：
- 同一教师同一时间只能存在一条有效课程
- 同一学员同一时间只能存在一条有效课程
- 状态为 `cancelled` 的课程不参与冲突校验

### 5.2 课时扣减规则
- 只有 `schedule.schedule_status = completed` 时允许扣课时
- `lesson_record.lesson_deducted = 1` 后不得重复扣减
- 请假和取消状态默认不扣减

### 5.3 课时包状态建议
- `pending`：未生效
- `active`：生效中
- `exhausted`：课时耗尽
- `expired`：已过期
- `closed`：人为关闭

### 5.4 审计建议
涉及以下动作时建议写入 `operation_log`：
- 新增/修改/删除学员
- 排课、调课、请假、取消、补课
- 课时扣减
- 缴费登记与退款

---

## 6. 后续优化建议

如需进入 V1.1，可继续扩展：
- 家长信息独立成表
- 多教师、多校区支持
- 通知提醒记录表
- 退款记录表
- 系统参数配置表（如无故缺课是否扣课时）
