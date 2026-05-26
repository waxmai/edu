import type { RouteRecordRaw } from 'vue-router'

export const scheduleRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '/org/schedules',
        name: 'OrgSchedule',
        component: () => import('@/modules/schedule/pages/OrgSchedulePage.vue'),
        meta: {
          title: '机构排课管理',
          permission: 'schedule:list',
          menuPermission: 'menu:schedule:view',
          roles: ['org_admin'],
        },
      },
      {
        path: '/campus/schedules',
        name: 'CampusSchedule',
        component: () => import('@/modules/schedule/pages/CampusSchedulePage.vue'),
        meta: {
          title: '校区排课管理',
          permission: 'schedule:list',
          menuPermission: 'menu:schedule:view',
          roles: ['campus_admin'],
        },
      },
      {
        path: '/my-schedules',
        name: 'MySchedule',
        component: () => import('@/modules/schedule/pages/MySchedulePage.vue'),
        meta: {
          title: '我的排课',
          permission: 'schedule:list',
          menuPermission: 'menu:schedule:view',
          roles: ['teacher'],
        },
      },
      {
        path: '/org/lesson-packages',
        name: 'OrgLessonPackage',
        component: () => import('@/modules/lesson-package/pages/OrgLessonPackagePage.vue'),
        meta: {
          title: '机构课时包管理',
          permission: 'lesson-package:list',
          menuPermission: 'menu:lesson-package:view',
          roles: ['org_admin'],
        },
      },
      {
        path: '/campus/lesson-packages',
        name: 'CampusLessonPackage',
        component: () => import('@/modules/lesson-package/pages/CampusLessonPackagePage.vue'),
        meta: {
          title: '校区课时包管理',
          permission: 'lesson-package:list',
          menuPermission: 'menu:lesson-package:view',
          roles: ['campus_admin'],
        },
      },
      {
        path: '/org/lesson-records',
        name: 'OrgLessonRecord',
        component: () => import('@/modules/lesson-record/pages/OrgLessonRecordPage.vue'),
        meta: {
          title: '机构上课记录管理',
          permission: 'lesson-record:list',
          menuPermission: 'menu:lesson-record:view',
          roles: ['org_admin'],
        },
      },
      {
        path: '/campus/lesson-records',
        name: 'CampusLessonRecord',
        component: () => import('@/modules/lesson-record/pages/CampusLessonRecordPage.vue'),
        meta: {
          title: '校区上课记录管理',
          permission: 'lesson-record:list',
          menuPermission: 'menu:lesson-record:view',
          roles: ['campus_admin'],
        },
      },
      {
        path: '/my-lesson-records',
        name: 'MyLessonRecord',
        component: () => import('@/modules/lesson-record/pages/MyLessonRecordPage.vue'),
        meta: {
          title: '我的上课记录',
          permission: 'lesson-record:list',
          menuPermission: 'menu:lesson-record:view',
          roles: ['teacher'],
        },
      },
      {
        path: '/org/reschedules',
        name: 'OrgReschedule',
        component: () => import('@/modules/reschedule/pages/OrgReschedulePage.vue'),
        meta: {
          title: '机构调补课管理',
          permission: 'reschedule:list',
          menuPermission: 'menu:reschedule:view',
          roles: ['org_admin'],
        },
      },
      {
        path: '/campus/reschedules',
        name: 'CampusReschedule',
        component: () => import('@/modules/reschedule/pages/CampusReschedulePage.vue'),
        meta: {
          title: '校区调补课管理',
          permission: 'reschedule:list',
          menuPermission: 'menu:reschedule:view',
          roles: ['campus_admin'],
        },
      },
      {
        path: '/my-reschedules',
        name: 'MyReschedule',
        component: () => import('@/modules/reschedule/pages/MyReschedulePage.vue'),
        meta: {
          title: '我的调课申请',
          permission: 'reschedule:list',
          menuPermission: 'menu:reschedule:view',
          roles: ['teacher'],
        },
      },
      {
        path: '/courses',
        name: 'Course',
        component: () => import('@/modules/course/pages/CourseListPage.vue'),
        meta: {
          title: '课程定义',
          permission: 'course:list',
          menuPermission: 'menu:course:view',
          roles: ['org_admin', 'campus_admin'],
        },
      },
    ],
  },
]
