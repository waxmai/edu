# 前端 Web 模块设计说明

## 1. 文档说明

本文档用于明确 `frontend/web` 的前端工程模块划分、路由结构、页面清单和基础协作规范，作为后续前端脚手架初始化、页面开发和前后端联调的基线文档。

适用范围：
- 教培管理后台 Web 端
- 管理员 / 教师使用场景
- Vue 3 + TypeScript + Element Plus 技术栈

---

## 2. 前端定位

`frontend/web` 用于承载当前教培系统的后台管理端，主要服务以下业务场景：
- 学员信息维护
- 排课与调补课处理
- 课时包与收费管理
- 上课记录录入
- 仪表盘数据查看
- 系统用户管理

该端以中后台操作效率为核心，不强调营销展示，优先保证：
- 信息密度合理
- 表格、表单、弹窗交互顺畅
- 路径清晰，低学习成本

---

## 3. 技术栈建议

推荐技术栈：
- Vue 3
- TypeScript
- Vite
- Vue Router
- Pinia
- Element Plus
- Axios
- ECharts
- UnoCSS（可选）

设计原则：
- 优先稳定、易维护
- 优先适合后台表单和表格场景
- 优先与后端 RESTful API 直连

---

## 4. 前端目录结构建议

```text
frontend/web/
├── public/
├── src/
│   ├── api/
│   ├── assets/
│   ├── components/
│   │   ├── common/
│   │   └── business/
│   ├── layouts/
│   ├── router/
│   │   ├── index.ts
│   │   └── routes/
│   ├── stores/
│   ├── modules/
│   ├── styles/
│   ├── types/
│   ├── utils/
│   ├── views/
│   ├── App.vue
│   └── main.ts
├── .env.development
├── .env.production
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

---

## 5. 模块划分说明

## 5.1 模块划分原则

前端按业务域拆分，而不是简单按页面堆放。每个模块尽量形成：
- 页面
- 局部组件
- 类型定义
- 状态常量
- 组合式逻辑

这样有利于：
- 和后端业务模块保持一致
- 后续按业务调整更快定位
- 减少跨模块污染

---

## 5.2 模块清单

### auth
负责登录认证、token 管理、当前用户信息、路由鉴权。

### dashboard
负责首页仪表盘、统计卡片、提醒信息和最近课程展示。

### student
负责学员列表、学员详情、学员新增/编辑、状态管理。

### course
负责课程定义、标准时长、标准收费等基础课程配置。

### lesson-package
负责课时包列表、详情、剩余课时展示、低课时提醒。

### schedule
负责排课日历、排课新增/编辑、冲突提示、课程状态展示。

### lesson-record
负责上课记录列表、上课内容录入、作业和反馈维护。

### payment
负责缴费记录、报名缴费、续费缴费、欠费展示和收入入口。

### reschedule
负责请假登记、调课处理、待补课列表、补课安排和历史记录。

### system
负责用户管理、角色权限、系统参数等后台能力。

---

## 6. 模块内部建议结构

以 `student` 模块为例：

```text
modules/student/
├── pages/
│   ├── StudentListPage.vue
│   └── StudentDetailPage.vue
├── components/
│   ├── StudentFormDialog.vue
│   ├── StudentSearchForm.vue
│   └── StudentStatusTag.vue
├── composables/
│   ├── useStudentList.ts
│   └── useStudentDetail.ts
├── constants/
│   └── student.ts
└── types/
    └── student.ts
```

以 `schedule` 模块为例：

```text
modules/schedule/
├── pages/
│   └── ScheduleCalendarPage.vue
├── components/
│   ├── ScheduleFormDialog.vue
│   ├── ScheduleDetailDrawer.vue
│   └── ScheduleStatusTag.vue
├── composables/
│   └── useScheduleCalendar.ts
├── constants/
│   └── schedule.ts
└── types/
    └── schedule.ts
```

---

## 7. 路由设计

## 7.1 路由设计原则

- 一级菜单对应一级业务模块
- 列表页优先，详情页作为二级页
- 不做过深嵌套路由，避免维护复杂度上升

---

## 7.2 一级菜单建议

- 首页
- 学员管理
- 课时包管理
- 排课管理
- 上课记录
- 收费管理
- 调补课管理
- 系统管理

---

## 7.3 路由清单建议

```text
/login
/
  /dashboard
  /students
  /students/:id
  /courses
  /lesson-packages
  /schedules
  /lesson-records
  /payments
  /reschedules
  /system/users
```

### 路由说明

#### `/login`
登录页。

#### `/dashboard`
首页仪表盘。

#### `/students`
学员列表页。

#### `/students/:id`
学员详情页。

#### `/courses`
课程定义管理页。

#### `/lesson-packages`
课时包管理页。

#### `/schedules`
排课管理页。

#### `/lesson-records`
上课记录页。

#### `/payments`
收费管理页。

#### `/reschedules`
调补课管理页。

#### `/system/users`
用户管理页。

---

## 8. 页面清单设计

## 8.1 第一批核心页面

建议优先开发以下页面，用于打通最小业务闭环：

1. 登录页
2. 首页仪表盘
3. 学员列表页
4. 学员详情页
5. 课时包管理页
6. 排课管理页
7. 上课记录页
8. 收费管理页

---

## 8.2 第二批页面

在第一批稳定后补充：

9. 调补课管理页
10. 课程定义页
11. 用户管理页
12. 统计分析页

---

## 8.3 页面职责说明

### 登录页
- 用户名密码输入
- 登录提交
- 登录失败提示

### 首页仪表盘
- 今日课程数
- 本周课程数
- 待补课数
- 低课时提醒人数
- 本月收入
- 最近课程安排

### 学员列表页
- 搜索筛选
- 新增学员
- 学员列表展示
- 快速进入详情/排课/缴费

### 学员详情页
- 基础信息
- 课时包信息
- 缴费记录
- 上课记录
- 调补课记录

### 课时包管理页
- 课时包列表
- 剩余课时高亮
- 新增课时包
- 编辑课时包

### 排课管理页
- 日历课表
- 快速新增排课
- 课程详情弹窗
- 冲突提示

### 上课记录页
- 上课记录列表
- 标记课程完成
- 录入课堂内容、作业、反馈

### 收费管理页
- 缴费记录列表
- 新增报名缴费
- 新增续费缴费
- 欠费高亮

### 调补课管理页
- 待补课列表
- 请假登记
- 调课处理
- 补课安排
- 历史记录查询

### 用户管理页
- 用户列表
- 新增用户
- 启用/禁用账号

---

## 9. 状态管理建议

## 9.1 全局 Store 建议

### `stores/auth.ts`
负责：
- token
- 当前用户信息
- 登录状态
- 登出
- 路由鉴权基础能力

### `stores/app.ts`
负责：
- 菜单折叠状态
- 全局 loading
- 主题和布局参数
- 系统级字典缓存

---

## 9.2 业务数据管理建议

业务列表数据不建议全部塞入 Pinia。推荐：
- 跨页面共享状态放 Pinia
- 页面局部状态放模块内 composable
- 查询条件、分页、弹窗开关优先模块内管理

---

## 10. 接口协作建议

## 10.1 API 目录规划

建议按模块拆分：

```text
api/
  auth.ts
  dashboard.ts
  student.ts
  course.ts
  lessonPackage.ts
  schedule.ts
  lessonRecord.ts
  payment.ts
  reschedule.ts
  system.ts
```

---

## 10.2 接口设计原则

- 命名尽量与后端模块一致
- 参数类型和返回类型独立定义
- 不在页面中直接拼接原始请求逻辑
- 接口异常统一在请求层做基础处理

---

## 11. 与后端模块映射关系

| 前端模块 | 后端模块 | 主要文档参考 |
| --- | --- | --- |
| student | student | docs/business/student.md |
| lesson-package | lesson_package | docs/business/lesson-package.md |
| schedule | schedule | docs/business/schedule.md |
| payment | payment_record | docs/business/payment.md |
| reschedule | reschedule_record | docs/business/reschedule.md |

---

## 12. 开发顺序建议

### 第一阶段：最小闭环
1. auth
2. dashboard
3. student
4. lesson-package
5. schedule
6. lesson-record
7. payment

### 第二阶段：教务细节补齐
8. reschedule
9. course
10. system

### 第三阶段：增强体验
11. 统计分析
12. 提醒能力
13. 高级筛选与批量操作

---

## 13. 结论

`frontend/web` 应按业务模块拆分，优先围绕后台管理端最小业务闭环建设。当前最适合的开发路径不是直接堆页面，而是先固定模块边界、路由结构和页面职责，再初始化前端脚手架并逐步接入后端接口。
