import type { RouteLocationNormalizedLoaded, RouteMeta } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export interface PermissionGuardOptions {
  roles?: string[]
  permission?: string
  menuPermission?: string
  featureFlag?: string
}

export interface MenuItem extends PermissionGuardOptions {
  to: string
  label: string
}

export function canAccess(options: PermissionGuardOptions = {}) {
  const authStore = useAuthStore()
  const roles = options.roles || []
  return (
    authStore.hasAnyRole(roles) &&
    authStore.hasPermission(options.permission) &&
    authStore.hasMenuPermission(options.menuPermission) &&
    authStore.hasFeatureFlag(options.featureFlag)
  )
}

export function canAccessMeta(meta: RouteMeta) {
  return canAccess({
    roles: Array.isArray(meta.roles) ? (meta.roles as string[]) : [],
    permission: typeof meta.permission === 'string' ? meta.permission : '',
    menuPermission: typeof meta.menuPermission === 'string' ? meta.menuPermission : '',
    featureFlag: typeof meta.featureFlag === 'string' ? meta.featureFlag : '',
  })
}

export function canAccessRoute(route: Pick<RouteLocationNormalizedLoaded, 'matched'>) {
  return route.matched.every((record) => canAccessMeta(record.meta))
}

export function filterMenuItems<T extends MenuItem>(items: T[]) {
  return items.filter((item) => canAccess(item))
}

export function usePermission() {
  return {
    can: canAccess,
    canAccess,
    canAccessMeta,
    canAccessRoute,
    filterMenuItems,
  }
}
