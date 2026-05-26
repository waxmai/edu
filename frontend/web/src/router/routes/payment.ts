import type { RouteRecordRaw } from 'vue-router'

export const paymentRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    children: [
      {
        path: '/org/payments',
        name: 'OrgPayment',
        component: () => import('@/modules/payment/pages/OrgPaymentPage.vue'),
        meta: {
          title: '机构收费管理',
          permission: 'payment:list',
          menuPermission: 'menu:payment:view',
          roles: ['org_admin'],
        },
      },
      {
        path: '/campus/payments',
        name: 'CampusPayment',
        component: () => import('@/modules/payment/pages/CampusPaymentPage.vue'),
        meta: {
          title: '校区收费管理',
          permission: 'payment:list',
          menuPermission: 'menu:payment:view',
          roles: ['campus_admin'],
        },
      },
    ],
  },
]
