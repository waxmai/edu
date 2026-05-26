import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocationNormalized } from 'vue-router'
import { authRoutes } from './routes/auth'
import { dashboardRoutes } from './routes/dashboard'
import { studentRoutes } from './routes/student'
import { scheduleRoutes } from './routes/schedule'
import { paymentRoutes } from './routes/payment'
import { systemRoutes } from './routes/system'
import { subscriptionRoutes } from './routes/subscription'
import { useAuthStore } from '@/stores/auth'
import { syncForcedPasswordDialog } from '@/stores/password-dialog'
import { canAccessMeta } from '@/composables/usePermission'
import { clearRuntimeError, saveRuntimeError, summarizeUnknownError } from '@/utils/runtime-error'

const legacyRoleRedirects: Record<string, Record<string, string>> = {
  org_admin: {
    '/payments': '/org/payments',
    '/lesson-packages': '/org/lesson-packages',
    '/schedules': '/org/schedules',
    '/reschedules': '/org/reschedules',
    '/lesson-records': '/org/lesson-records',
    '/system/users': '/org/users',
  },
  campus_admin: {
    '/payments': '/campus/payments',
    '/lesson-packages': '/campus/lesson-packages',
    '/schedules': '/campus/schedules',
    '/reschedules': '/campus/reschedules',
    '/lesson-records': '/campus/lesson-records',
    '/system/users': '/campus/users',
  },
  platform_admin: {
    '/system/users': '/platform/users',
  },
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...authRoutes,
    ...dashboardRoutes,
    ...studentRoutes,
    ...scheduleRoutes,
    ...paymentRoutes,
    ...systemRoutes,
    ...subscriptionRoutes,
  ],
})

router.beforeEach((to: RouteLocationNormalized) => {
  const authStore = useAuthStore()
  const isPublic = Boolean(to.meta.public)
  const role = authStore.currentUser?.roleCode || ''

  if (!isPublic && !authStore.isAuthenticated) {
    return {
      path: '/login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if (to.path === '/login' && authStore.isAuthenticated) {
    return '/dashboard'
  }

  if (authStore.isAuthenticated && authStore.mustChangePassword) {
    authStore.activateForcePasswordChange()
    syncForcedPasswordDialog(true)
  }

  const roleRedirect = legacyRoleRedirects[role]?.[to.path]
  if (roleRedirect) {
    return roleRedirect
  }

  if (!isPublic) {
    for (const record of to.matched) {
      if (!canAccessMeta(record.meta)) {
        return '/dashboard'
      }
    }
  }

  return true
})

router.afterEach((to) => {
  clearRuntimeError()
  console.info('[router][afterEach]', to.fullPath)
})

router.onError((error, to) => {
  const detail = summarizeUnknownError(error)
  const routePath = to.fullPath || window.location.pathname || '/'
  saveRuntimeError({
    scope: 'router',
    route: routePath,
    message: detail,
    time: new Date().toISOString(),
  })
  console.error('[router][error]', routePath, error)
})

export default router
