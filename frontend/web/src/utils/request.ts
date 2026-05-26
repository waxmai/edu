import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { syncForcedPasswordDialog } from '@/stores/password-dialog'

const resolvedApiBaseURL = (import.meta.env.VITE_API_BASE_URL || '/api/v1').trim()

const request = axios.create({
  baseURL: resolvedApiBaseURL,
  timeout: 10000,
})

request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  if (authStore.refreshToken) {
    config.headers['X-Refresh-Token'] = authStore.refreshToken
  }
  return config
})

request.interceptors.response.use(
  (response) => response,
  async (error) => {
    const authStore = useAuthStore()
    const originalRequest = error?.config
    if (error?.response?.status === 401 && !originalRequest?._retried && authStore.refreshToken) {
      originalRequest._retried = true
      const refreshed = await authStore.tryRefresh()
      if (refreshed) {
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = `Bearer ${authStore.token}`
        originalRequest.headers['X-Refresh-Token'] = authStore.refreshToken
        return request(originalRequest)
      }
    }
    if (error?.response?.status === 401) {
      authStore.clearAuth()
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        const redirect = `${window.location.pathname}${window.location.search}${window.location.hash}`
        window.location.href = `/login?redirect=${encodeURIComponent(redirect)}`
      }
    } else if (error?.response?.status === 403) {
      const message = error?.response?.data?.message || ''
      if (authStore.isAuthenticated && authStore.mustChangePassword && message.includes('请先修改密码')) {
        authStore.activateForcePasswordChange()
        syncForcedPasswordDialog(true)
      }
    }
    return Promise.reject(error)
  },
)

export default request
