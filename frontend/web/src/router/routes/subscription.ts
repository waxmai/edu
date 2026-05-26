import type { RouteRecordRaw } from 'vue-router'

export const subscriptionRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '/org/subscription',
        name: 'OrgSubscription',
        component: () => import('@/modules/subscription/pages/OrgSubscriptionPage.vue'),
        meta: {
          title: '订阅中心',
          roles: ['org_admin'],
          permission: 'auth:me:view',
          menuPermission: 'menu:subscription:center:view',
        },
      },
      {
        path: '/platform/subscriptions',
        name: 'PlatformSubscriptions',
        component: () => import('@/modules/subscription/pages/PlatformSubscriptionPage.vue'),
        meta: {
          title: '续费跟进',
          roles: ['platform_admin', 'platform_ops', 'platform_finance'],
          permission: 'platform:subscription:list',
          menuPermission: 'menu:platform:view',
          featureFlag: 'subscription_center',
        },
      },
    ],
  },
]
