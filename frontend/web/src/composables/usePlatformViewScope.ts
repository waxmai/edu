import { useAuthStore } from '@/stores/auth'

export function usePlatformViewScope() {
  const authStore = useAuthStore()

  function withPlatformViewScope<T extends Record<string, unknown>>(params: T): T {
    if (!authStore.hasPlatformViewAsScope()) {
      return params
    }
    return {
      ...params,
      organizationId: authStore.platformViewAs?.organizationId,
      campusId: authStore.platformViewAs?.campusId,
    }
  }

  return { withPlatformViewScope }
}
