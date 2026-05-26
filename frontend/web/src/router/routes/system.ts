import type { RouteRecordRaw } from 'vue-router'

export const systemRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '/platform/users',
        name: 'PlatformUsers',
        component: () => import('@/modules/system/pages/SystemUserListPage.vue'),
        meta: {
          title: '平台用户',
          roles: ['platform_admin'],
          permission: 'system:user:list',
          menuPermission: 'menu:system:user:view',
        },
      },
      {
        path: '/org/users',
        name: 'OrgUsers',
        component: () => import('@/modules/system/pages/OrgUserPage.vue'),
        meta: {
          title: '机构用户管理',
          roles: ['org_admin'],
          permission: 'system:user:list',
          menuPermission: 'menu:system:user:view',
          featureFlag: 'user_management',
        },
      },
      {
        path: '/org/settings',
        name: 'OrganizationSettings',
        component: () => import('@/modules/system/pages/OrganizationSettingsPage.vue'),
        meta: {
          title: '机构设置',
          roles: ['org_admin'],
          permission: 'tenant:settings:view',
          menuPermission: 'menu:dashboard:view',
        },
      },
      {
        path: '/campus/users',
        name: 'CampusUsers',
        component: () => import('@/modules/system/pages/CampusUserPage.vue'),
        meta: {
          title: '校区用户管理',
          roles: ['campus_admin'],
          permission: 'system:user:list',
          menuPermission: 'menu:system:user:view',
          featureFlag: 'user_management',
        },
      },
      {
        path: '/platform/tenants',
        name: 'PlatformTenants',
        component: () => import('@/modules/system/pages/PlatformOrganizationPage.vue'),
        meta: {
          title: '机构管理',
          roles: ['platform_admin', 'platform_ops', 'platform_finance', 'platform_support'],
          permission: 'platform:tenant:list',
          menuPermission: 'menu:platform:orgs:view',
        },
      },
      {
        path: '/platform/ops',
        name: 'PlatformOperations',
        component: () => import('@/modules/system/pages/PlatformOperationsPage.vue'),
        meta: {
          title: '平台运营总览',
          roles: ['platform_admin', 'platform_ops'],
          permission: 'platform:ops:view',
          menuPermission: 'menu:platform:view',
        },
      },
      {
        path: '/platform/audit-logs',
        name: 'PlatformAuditLogs',
        component: () => import('@/modules/system/pages/PlatformAuditLogPage.vue'),
        meta: {
          title: '审计日志',
          roles: ['platform_admin', 'platform_auditor'],
          permission: 'platform:audit:list',
          menuPermission: 'menu:platform:view',
          featureFlag: 'audit_export',
        },
      },
      {
        path: '/platform/recovery-alerts',
        name: 'PlatformRecoveryAlerts',
        component: () => import('@/modules/system/pages/PlatformRecoveryAlertPage.vue'),
        meta: {
          title: '恢复失败告警',
          roles: ['platform_admin', 'platform_auditor', 'platform_support'],
          permission: 'platform:recovery-alert:list',
          menuPermission: 'menu:platform:view',
          featureFlag: 'recovery_ops',
        },
      },
      {
        path: '/platform/campuses',
        name: 'PlatformCampuses',
        component: () => import('@/modules/system/pages/PlatformCampusPage.vue'),
        meta: {
          title: '校区管理',
          roles: ['platform_admin', 'platform_ops', 'platform_support', 'org_admin'],
          permission: 'platform:campus:list',
          menuPermission: 'menu:campus:manage:view',
        },
      },
    ],
  },
]
