import type { RouteRecordRaw } from 'vue-router'

export const studentRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '/students',
        name: 'StudentList',
        component: () => import('@/modules/student/pages/StudentListPage.vue'),
        meta: {
          title: '学员管理',
          permission: 'student:list',
          menuPermission: 'menu:student:view',
          roles: ['org_admin', 'campus_admin'],
        },
      },
      {
        path: '/my-students',
        name: 'MyStudentList',
        component: () => import('@/modules/student/pages/MyStudentListPage.vue'),
        meta: {
          title: '我的学生',
          permission: 'student:list',
          menuPermission: 'menu:student:view',
          roles: ['teacher'],
        },
      },
      {
        path: '/students/:id',
        name: 'StudentDetail',
        component: () => import('@/modules/student/pages/StudentDetailPage.vue'),
        meta: {
          title: '学员详情',
          permission: 'student:get',
          menuPermission: 'menu:student:view',
          roles: ['org_admin', 'campus_admin'],
        },
      },
      {
        path: '/my-students/:id',
        name: 'MyStudentDetail',
        component: () => import('@/modules/student/pages/MyStudentDetailPage.vue'),
        meta: {
          title: '我的学生详情',
          permission: 'student:get',
          menuPermission: 'menu:student:view',
          roles: ['teacher'],
        },
      },
    ],
  },
]
